package proxy

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"time"
)

func ProxyHandler(pool LoadBalancer, timeout time.Duration) http.HandlerFunc{

	return func(w http.ResponseWriter,r * http.Request){
		backend := pool.GetNextValidPeer()

		if backend == nil{
			http.Error(w, "503 Service unavailable",http.StatusServiceUnavailable)
			return
		}

		backend.IncrementConnections()

		defer backend.DecrementConnections()

		proxy := httputil.NewSingleHostReverseProxy(backend.URL)

		ctx, cancel := context.WithTimeout(r.Context(),timeout)

		defer cancel()
		r = r.WithContext(ctx)

		proxy.ErrorHandler =  func(w http.ResponseWriter, r *http.Request,err error){
			log.Println("Backend ", backend.URL.String()," failed: ",err)
			backend.SetAlive(false)
			http.Error(w, "502 Bad Gateway",http.StatusBadGateway)
		}

		proxy.ServeHTTP(w,r)
	}
}