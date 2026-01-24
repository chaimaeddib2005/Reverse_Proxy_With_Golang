package proxy

import "net/url"

type LoadBalancer interface {

	GetNextValidPeer() *Backend
		GetLeastConnBackend() *Backend
	AddBackend(backend *Backend)
	SetBackendStatus(uri *url.URL, alive bool)

}
