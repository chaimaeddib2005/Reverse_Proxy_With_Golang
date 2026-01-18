package main

import (
	"fmt"
	"net/http"
	"log"
	"reverseproxy.com/config"
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
	
	log.Fatal(http.ListenAndServe(":8080", nil))


}