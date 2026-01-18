package health

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"
	"reverseproxy.com/proxy"
)

type HealthChecker struct{
	pool   *proxy.ServerPool
	interval time.Duration
	timeout time.Duration
	method string
}

func NewHealthChecker(pool *proxy.ServerPool,interval,timeout time.Duration,method string) *HealthChecker{
	return &HealthChecker{
		pool: pool,
		interval: interval,
		timeout: timeout,
		method: method,
	}
}

func (hc *HealthChecker) Start(ctx context.Context){
	ticker :=  time.NewTicker(hc.interval)
	defer ticker.Stop()

	log.Printf("Health Checker started (interval : %v, timeout: %v)",hc.interval,hc.timeout)

	hc.checkAllBackends()

	for{
		select {
        case <-ticker.C:
            hc.checkAllBackends()
        case <-ctx.Done():
            log.Println("Health checker stopped")
            return
        }
	}
}

func (hc *HealthChecker) checkAllBackends(){
	for _,backend := range hc.pool.Backends{
		go hc.checkBackend(backend)
	}
}

func (hc *HealthChecker) checkBackend(backend *proxy.Backend){

	wasAlive := backend.IsAlive()

	isAlive := hc.isBackendAlive(backend)

	backend.SetAlive(isAlive)

	if wasAlive && !isAlive{
		log.Printf("Backend %s is Down",backend.URL.String())
	}else if !wasAlive && isAlive{
		log.Printf("Backend %s Up", backend.URL.String())
	}
}


func (hc *HealthChecker) isBackendAlive(backend *proxy.Backend) bool{
	  if hc.method == "tcp"{
		return hc.checkTCP(backend)
	  }else{
		return hc.checkHTTP(backend)
	  }
}

func (hc *HealthChecker) checkTCP(backend *proxy.Backend) bool{
	ctx, cancel := context.WithTimeout(context.Background(),hc.timeout)
	defer cancel()

	var dialer net.Dialer
	conn,err := dialer.DialContext(ctx,"tcp",backend.URL.Host)
	if err != nil{
		return false
	}
	defer conn.Close()

	return true

}

func (hc *HealthChecker) checkHTTP(backend *proxy.Backend) bool{
	ctx, cancel := context.WithTimeout(context.Background(),hc.timeout)
	defer cancel()
	healthURL :=  backend.URL.String()+"/health"
	
	req, err := http.NewRequestWithContext(ctx,"GET",healthURL,nil)
	if err != nil{
		return false
	}

	client := &http.Client{
		Timeout: hc.timeout,
	}
	resp, err  := client.Do(req)
	if err != nil{
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode < 500

}