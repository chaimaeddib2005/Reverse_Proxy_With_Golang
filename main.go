package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
	"reverseproxy.com/config"
	"reverseproxy.com/health"
	"reverseproxy.com/proxy"
	"reverseproxy.com/admin"
)

func main(){
	configuration, err := config.LoadConfiguration()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}
	
	pool := &proxy.ServerPool{}
	
	for _, backendConfig := range configuration.BackendsConfig {
		parsedURL, err := url.Parse(backendConfig.URL)
		if err != nil {
			log.Printf("Invalid backend URL %s: %v", backendConfig.URL, err)
			continue
		}
		
		weight := backendConfig.Weight
		if weight == 0 {
			weight = 1
		}
		
		backend := &proxy.Backend{
			URL:          parsedURL,
			Alive:        true,
			CurrentConns: 0,
			Weight:       weight,
		}
		
		pool.AddBackend(backend)
		fmt.Printf("Added backend: %s (weight: %d)\n", parsedURL, weight)
	}

	fmt.Println("The number of backend servers is:", len(pool.Backends))
	
	var loadBalancer proxy.LoadBalancer
	
	if configuration.EnableStickySessions {
		stickyTTL := configuration.StickySessionTTL
		if stickyTTL == 0 {
			stickyTTL = 30 * time.Minute
		}
		loadBalancer = proxy.NewStickySessionPool(pool, stickyTTL)
		fmt.Println("Sticky sessions enabled with TTL:", stickyTTL)
	} else {
		loadBalancer = pool
		if configuration.Strategy == "least-conn" {
			fmt.Println("Using least-connections load balancing")
		} else {
			fmt.Println("Using weighted round-robin load balancing")
		}
	}
	
	fmt.Printf("Proxy server starting on :%d\n", configuration.Port)

	http.HandleFunc("/", proxy.ProxyHandler(loadBalancer, configuration.Backend_timeout, configuration.EnableStickySessions, configuration.Strategy))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	healthChecker := health.NewHealthChecker(pool, configuration.HealthCheckFreq, configuration.Backend_timeout, configuration.HealthCheckMethod)
	adminAPI := admin.NewAdminAPI(pool)
	
	adminMux := http.NewServeMux()
	adminAPI.SetUpRoutes(adminMux)
	
	adminServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", configuration.Admin_port),
		Handler: adminMux,
	}

	go healthChecker.Start(ctx)

	proxyServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", configuration.Port),
		Handler: nil,
	}
	
	go func() {
		log.Printf("Proxy server listening on :%d\n", configuration.Port)
		if err := proxyServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Proxy server error: %v", err)
		}
	}()
	
	go func() {
		log.Println("Admin API listening on", adminServer.Addr)
		if err := adminServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Admin server error: %v", err)
		}
	}()
	
	waitForShutdown(cancel, proxyServer, adminServer)
}

func waitForShutdown(cancel context.CancelFunc, server *http.Server, adminServer *http.Server) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	
	log.Printf("Shutdown signal received")
	cancel()
	
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Proxy Server shutdown error: %v", err)
	} else {
		log.Println("Proxy server stopped gracefully")
	}
	
	if err := adminServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Admin Server shutdown error: %v", err)
	} else {
		log.Println("Admin server stopped gracefully")
	}
	
	log.Println("All servers stopped")
}