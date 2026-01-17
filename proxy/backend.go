package proxy

import (
	"net/url"
	"sync"
	"sync/atomic"
)

type Backend struct {

URL *url.URL `json:"url"`
Alive bool `json:"alive"`
CurrentConns int64 `json:"current_connections"`
mux sync.RWMutex

}

func (b *Backend) SetAlive(alive bool){
	b.mux.Lock()
	b.Alive = alive
	b.mux.Unlock()

}

func (b *Backend) IsAlive() bool{
	b.mux.RLock()
	r :=  b.Alive
	b.mux.RUnlock()
	return r
}

func (b *Backend) IncrementConnections(){
	b.mux.Lock()
	atomic.AddInt64(&b.CurrentConns,1)
	b.mux.Unlock()
}

func (b *Backend) DecrementConnections(){
	b.mux.Lock()
	atomic.AddInt64(&b.CurrentConns,1)
	b.mux.Unlock()
}

