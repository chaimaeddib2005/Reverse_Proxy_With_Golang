package proxy

import "net/http"

func ProxyHandler(pool LoadBalancer) http.HandlerFunc{

	return func(w http.ResponseWriter,r * http.Request){
		backend := pool.GetNextValidPeer()

		if backend == nil{
			http.Error(w, "Service unavailable",503)
			return
		}

		backend.IncrementConnections()

		defer backend.DecrementConnections()

		w.Write([]byte("Forwarding to: "+backend.URL.String()))
	}
}