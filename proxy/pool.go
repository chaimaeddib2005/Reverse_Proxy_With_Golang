package proxy


type ServerPool struct {

Backends []*Backend `json:"backends"`
Current uint64 `json:"current"` // Used for Round-Robin

}