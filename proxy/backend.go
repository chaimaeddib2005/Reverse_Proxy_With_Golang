package proxy

import (
	"net/url"
	"sync"
	"sync/atomic"
)

type Backend struct {
	URL          *url.URL `json:"url"`
	Alive        bool     `json:"alive"`
	CurrentConns int64    `json:"current_connections"`
	mux          sync.RWMutex
	Weight       int      `json:"weight"`
}

func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	b.Alive = alive
	b.mux.Unlock()
}

func (b *Backend) IsAlive() bool {
	b.mux.RLock()
	r := b.Alive
	b.mux.RUnlock()
	return r
}

func (b *Backend) IncrementConnections() {
	atomic.AddInt64(&b.CurrentConns, 1)
}

func (b *Backend) DecrementConnections() {
	atomic.AddInt64(&b.CurrentConns, -1)
}

func (b *Backend) GetCurrentConns() int64 {
	return atomic.LoadInt64(&b.CurrentConns)
}