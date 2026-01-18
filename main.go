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
	configuration ,err := config.LoadConfiguration()
	if err != nil{
		log.Fatalf("Configuration error: %v", err)
	}
	pool := proxy.ServerPool{}
	back := proxy.Backend{}
	
	for _,Url := range(configuration.BackendsUrls){
		back.URL = Url
		back.SetAlive(true)
		back.CurrentConns = 0
		pool.Backends = append(pool.Backends, &proxy.Backend{URL:Url,Alive: true,CurrentConns:  0})
	}

	fmt.Println("The numder of backend servers is: ",len(pool.Backends))
	
	fmt.Println("Proxy server starting on :8080")

	http.HandleFunc("/", proxy.ProxyHandler(&pool,configuration.Backend_timeout ))

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	healthChecker := health.NewHealthChecker(&pool,configuration.HealthCheckFreq,configuration.Backend_timeout,configuration.HealthCheckMethod)
	adminAPI := admin.NewAdminAPI(&pool)
	adminMux := http.NewServeMux()
	adminAPI.SetUpRoutes(adminMux)
	
	adminServer := &http.Server{
		Addr:    ":8081",
		Handler: adminMux,
	}

	go healthChecker.Start(ctx)

    go func(){
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()
	server := &http.Server{
        Addr:    ":8080",
        Handler: nil,
    }
	go func() {
		log.Println("Admin API listening on :8081")
		if err := adminServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Admin server error: %v", err)
		}
	}()
	waitForShutdown(ctx, cancel, server,adminServer)
}

func waitForShutdown(ctx context.Context, cancel context.CancelFunc, server *http.Server,adminServer *http.Server){

	sigChan := make(chan os.Signal,1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	log.Printf("Shutdown signal received")
	cancel()
	 shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 30*time.Second)
    defer shutdownCancel()
    
    if err := server.Shutdown(shutdownCtx); err != nil {
        log.Printf(" Proxy Server shutdown error: %v", err)
    }else if err := adminServer.Shutdown(shutdownCtx); err != nil{
		 log.Printf(" Admin Server shutdown error: %v", err)
	}
    
    log.Println("Servers stopped gracefully")


}