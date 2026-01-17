package proxy

import (
	"net/url"
	"sync"
	"sync/atomic"
)


type ServerPool struct {

Backends []*Backend `json:"backends"`
Current uint64 `json:"current"` // Used for Round-Robin
mux sync.RWMutex

}

func (p *ServerPool) GetNextValidPeer() *Backend {

    
    next := atomic.AddUint64(&p.Current, 1)

    
    p.mux.RLock()
    defer p.mux.RUnlock()

    if len(p.Backends) == 0 {
        return nil
    }

    start := int(next) % len(p.Backends)

    for i := 0; i < len(p.Backends); i++ {
        idx := (start + i) % len(p.Backends)

        backend := p.Backends[idx]

        backend.mux.RLock()
        alive := backend.Alive
        backend.mux.RUnlock()

        if alive {
            return backend
        }
    }
    return nil
}


func (p *ServerPool) AddBackend(backend *Backend){
	p.mux.Lock()
	p.Backends = append(p.Backends, backend)
	p.mux.Unlock()

}

func (p *ServerPool) SetBackendStatus(uri *url.URL, alive bool){
	for i := 0;i<len(p.Backends);i++{
		if p.Backends[i].URL == uri{
			p.Backends[i].SetAlive(alive)
			return
		}
	}
	
}

