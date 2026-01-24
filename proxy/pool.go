package proxy

import (
	"net/url"
	"sync"
	"sync/atomic"
)

type ServerPool struct {
	Backends      []*Backend `json:"backends"`
	Current       uint64     `json:"current"`
	Mux           sync.RWMutex
	CurrentWeight int
	MaxWeight     int
	GCD           int
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func (p *ServerPool) calculateGCD() int {
	if len(p.Backends) == 0 {
		return 1
	}

	result := p.Backends[0].Weight
	for i := 1; i < len(p.Backends); i++ {
		result = gcd(result, p.Backends[i].Weight)
	}
	return result
}

func (p *ServerPool) calculateMaxWeight() int {
	max := 0
	for _, backend := range p.Backends {
		if backend.Weight > max {
			max = backend.Weight
		}
	}
	return max
}

func (p *ServerPool) updateWeightMetrics() {
	p.GCD = p.calculateGCD()
	p.MaxWeight = p.calculateMaxWeight()
}

func (p *ServerPool) GetNextValidPeer() *Backend {
	p.Mux.RLock()
	defer p.Mux.RUnlock()

	if len(p.Backends) == 0 {
		return nil
	}

	allWeightsEqual := true
	firstWeight := p.Backends[0].Weight
	for _, b := range p.Backends {
		if b.Weight != firstWeight || b.Weight == 0 {
			allWeightsEqual = false
			break
		}
	}

	if allWeightsEqual && firstWeight == 0 {
		return p.simpleRoundRobin()
	}

	return p.weightedRoundRobin()
}

func (p *ServerPool) simpleRoundRobin() *Backend {
	next := atomic.AddUint64(&p.Current, 1)
	start := int(next) % len(p.Backends)

	for i := 0; i < len(p.Backends); i++ {
		idx := (start + i) % len(p.Backends)
		backend := p.Backends[idx]

		if backend.IsAlive() {
			return backend
		}
	}
	return nil
}

func (p *ServerPool) weightedRoundRobin() *Backend {
	totalWeight := 0
	var bestBackend *Backend
	maxEffectiveWeight := -1

	for _, backend := range p.Backends {
		if !backend.IsAlive() {
			continue
		}

		totalWeight += backend.Weight

		effectiveWeight := backend.Weight
		if bestBackend == nil || effectiveWeight > maxEffectiveWeight {
			bestBackend = backend
			maxEffectiveWeight = effectiveWeight
		}
	}

	if bestBackend == nil {
		return nil
	}

	return bestBackend
}

func (p *ServerPool) GetLeastConnBackend() *Backend {
	p.Mux.RLock()
	defer p.Mux.RUnlock()

	if len(p.Backends) == 0 {
		return nil
	}

	var selected *Backend
	minConns := int64(-1)

	for _, backend := range p.Backends {
		if !backend.IsAlive() {
			continue
		}

		conns := backend.GetCurrentConns()
		if selected == nil || conns < minConns {
			selected = backend
			minConns = conns
		}
	}

	return selected
}

func (p *ServerPool) AddBackend(backend *Backend) {
	p.Mux.Lock()

	if backend.Weight == 0 {
		backend.Weight = 1
	}

	p.Backends = append(p.Backends, backend)
	p.updateWeightMetrics()
	p.Mux.Unlock()
}

func (p *ServerPool) SetBackendStatus(uri *url.URL, alive bool) {
	p.Mux.Lock()
	defer p.Mux.Unlock()

	for i := 0; i < len(p.Backends); i++ {
		if p.Backends[i].URL.String() == uri.String() {
			p.Backends[i].SetAlive(alive)
			return
		}
	}
}