package admin

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"sync"
	"reverseproxy.com/proxy"
)


type AdminAPI struct{
	pool *proxy.ServerPool
	mux sync.RWMutex
}

func NewAdminAPI(pool *proxy.ServerPool) *AdminAPI{
	return &AdminAPI{
		pool:pool,
	}
}

func (a *AdminAPI) SetUpRoutes(mux *http.ServeMux){
	mux.HandleFunc("/status", a.handleStatus)
    mux.HandleFunc("/backends", a.handleBackends)
}

type StatusResponse struct{
	TotalBackends int `json:"total_backends"`
	ActiveBackends int `json:"active_backends"`
	Backends []BackendsStatus `json:"backends"`
}

type BackendsStatus struct{
	URL string `json:"url"`
	Alive bool `jon:"alive"`
	CurrentConnections int64 `json:"current_connections"`
}

type AddBackendsRequest struct{
	URL string `json:"url"`
}

type DeleteBackendsRequest struct{
	URL string `json:"url"`
}
func (a *AdminAPI) handleStatus(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodGet{
		http.Error(w, "Method not allowed",http.StatusMethodNotAllowed)
		return
	}
	a.mux.RLock()
	defer a.mux.Unlock()

	var backends []BackendsStatus
	activeCount := 0

	for _,backend :=  range a.pool.Backends{
		isAlive := backend.IsAlive()

		status := BackendsStatus{
			URL: backend.URL.String(),
			Alive: isAlive,
			CurrentConnections: backend.CurrentConns,
		}
		backends = append(backends, status)
		if isAlive{
			activeCount++
		}
	}

	response := StatusResponse{
		TotalBackends: len(a.pool.Backends),
		ActiveBackends: activeCount,
		Backends: backends,
	}
	w.Header().Set("Content-Type","application/json")
	json.NewEncoder(w).Encode(response)
	log.Printf("Status requested: %d/%d backends alive", activeCount,len(a.pool.Backends))
}


func (a *AdminAPI) handleBackends( w http.ResponseWriter, r *http.Request){
	switch r.Method{
	case http.MethodPost:
		a.handleAddBackend(w,r)
	case http.MethodDelete:
		a.handleDeleteBackend(w,r)
	default:
		http.Error(w,"Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *AdminAPI) handleAddBackend(w http.ResponseWriter,r *http.Request){
	var req AddBackendsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil{
		http.Error(w, "Invalid Json",http.StatusBadRequest)
		return
	}

	parsedURL, err := url.Parse(req.URL)
	 if err != nil{
		http.Error(w,"Invalid URL", http.StatusBadRequest)
	 }

	 a.mux.RUnlock()
	 newBackend := proxy.Backend{
		URL:          parsedURL,
        Alive:        true,
        CurrentConns: 0,
	 }

	 a.mux.Lock()
	 a.pool.Backends = append(a.pool.Backends,&newBackend)
	 a.mux.Unlock()

	 log.Printf("Backend added: %s", parsedURL.String())

	 w.Header().Set("Content-Type","application/json")

	 w.WriteHeader(http.StatusCreated)
	 json.NewEncoder(w).Encode(map[string]string{
		"message": "Backend added successfully",
        "url":     parsedURL.String(),
	 })


}

func (a *AdminAPI) handleDeleteBackend(w http.ResponseWriter,r *http.Request){
	var req DeleteBackendsRequest

	if err := json.NewDecoder(r.Body).Decode(&req);err != nil{
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
	}

	parsedURL,err := url.Parse(req.URL)
	if err != nil{
		http.Error(w, "Invalid URL", http.StatusBadRequest)
        return
	} 

	a.mux.Lock()
	defer a.mux.Unlock()

	found := false
	newBackends := make([]*proxy.Backend,0,len(a.pool.Backends))

	for _,backend := range a.pool.Backends{
		if backend.URL.String() ==  parsedURL.String(){
			found = true
			log.Printf("Backed removed: %s (had %d active connections)",
		backend.URL.String(),backend.CurrentConns)
		}else{
			newBackends = append(newBackends, backend)
		}
	}

	if !found{
		http.Error(w,"Backend not found", http.StatusNotFound)
		return
	}

	a.pool.Backends = newBackends

	w.Header().Set("Content-Type","application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Backend removed successfully",
        "url":     parsedURL.String(),
	})
}