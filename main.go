package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
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
	
	for _, Url := range configuration.BackendsUrls {
		pool.Backends = append(pool.Backends, &proxy.Backend{
			URL:          Url,
			Alive:        true,
			CurrentConns: 0,
		})
	}

	fmt.Println("The number of backend servers is: ", len(pool.Backends))
	
	
	var loadBalancer proxy.LoadBalancer
	
	if configuration.EnableStickySessions {
		
		stickyTTL := configuration.StickySessionTTL
		loadBalancer = proxy.NewStickySessionPool(pool, stickyTTL)
		fmt.Println("Sticky sessions enabled with TTL:", stickyTTL)
	} else {
		loadBalancer = pool
		fmt.Println("Using round-robin load balancing")
	}
	
	fmt.Println("Proxy server starting on :8080")

	http.HandleFunc("/", proxy.ProxyHandler(loadBalancer, configuration.Backend_timeout, configuration.EnableStickySessions))

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
		Addr:    ":8080",
		Handler: nil, 
	}
	
	go func() {
		log.Println("Proxy server listening on :8080")
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