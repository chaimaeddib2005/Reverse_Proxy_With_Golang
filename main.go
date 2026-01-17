package main

import (
	"fmt"
	

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
	
	//handler := proxy.ProxyHandler(&pool)
	for range(100){
		go func(){
			b1 := pool.GetNextValidPeer()
			fmt.Println(b1.URL.String())
			}()
	}
	



}