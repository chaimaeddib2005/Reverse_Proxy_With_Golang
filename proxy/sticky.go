package proxy

import (
    "net"
    "net/http"
    "net/url"
    "sync"
    "time"
)

type StickySession struct {
    Backend  *Backend
    LastSeen time.Time
}

type StickySessionPool struct {
    pool     *ServerPool
    sessions map[string]*StickySession
    mux      sync.RWMutex
    ttl      time.Duration
}

func NewStickySessionPool(pool *ServerPool, ttl time.Duration) *StickySessionPool {
    sp := &StickySessionPool{
        pool:     pool,
        sessions: make(map[string]*StickySession),
        ttl:      ttl,
    }
    
    go sp.cleanupExpiredSessions()
    
    return sp
}

func (sp *StickySessionPool) GetBackendForClient(r *http.Request) *Backend {
    clientIP := getClientIP(r)
    
    sp.mux.RLock()
    session, exists := sp.sessions[clientIP]
    sp.mux.RUnlock()
    
    if exists && session.Backend.IsAlive() {
        sp.mux.Lock()
        session.LastSeen = time.Now()
        sp.mux.Unlock()
        return session.Backend
    }
    
    backend := sp.pool.GetNextValidPeer()
    
    if backend != nil {
        sp.mux.Lock()
        sp.sessions[clientIP] = &StickySession{
            Backend:  backend,
            LastSeen: time.Now(),
        }
        sp.mux.Unlock()
    }
    
    return backend
}

func (sp *StickySessionPool) GetNextValidPeer() *Backend {
    return sp.pool.GetNextValidPeer()
}

func (sp *StickySessionPool) GetLeastConnBackend() *Backend {
    return sp.pool.GetLeastConnBackend()
}

func (sp *StickySessionPool) AddBackend(backend *Backend) {
    sp.pool.AddBackend(backend)
}

func (sp *StickySessionPool) SetBackendStatus(uri *url.URL, alive bool) {
    sp.pool.SetBackendStatus(uri, alive)
    
    if !alive {
        sp.mux.Lock()
        for ip, session := range sp.sessions {
            if session.Backend.URL.String() == uri.String() {
                delete(sp.sessions, ip)
            }
        }
        sp.mux.Unlock()
    }
}

func (sp *StickySessionPool) cleanupExpiredSessions() {
    ticker := time.NewTicker(sp.ttl)
    defer ticker.Stop()
    
    for range ticker.C {
        sp.mux.Lock()
        now := time.Now()
        for ip, session := range sp.sessions {
            if now.Sub(session.LastSeen) > sp.ttl {
                delete(sp.sessions, ip)
            }
        }
        sp.mux.Unlock()
    }
}

func getClientIP(r *http.Request) string {
    forwarded := r.Header.Get("X-Forwarded-For")
    if forwarded != "" {
        return forwarded
    }
    
    ip, _, _ := net.SplitHostPort(r.RemoteAddr)
    return ip
}