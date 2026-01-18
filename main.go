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
)

func main(){
	configuration ,err := config.LoadConfiguration()
	if err != nil{
		fmt.Println(err)
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

	go healthChecker.Start(ctx)

    go func(){
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()
	server := &http.Server{
        Addr:    ":8080",
        Handler: nil,
    }

	waitForShutdown(ctx, cancel, server)
}

func waitForShutdown(ctx context.Context, cancel context.CancelFunc, server *http.Server){

	sigChan := make(chan os.Signal,1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	log.Printf("Shutdown signal received")
	cancel()
	 shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 30*time.Second)
    defer shutdownCancel()
    
    if err := server.Shutdown(shutdownCtx); err != nil {
        log.Printf("Server shutdown error: %v", err)
    }
    
    log.Println("Server stopped gracefully")


}