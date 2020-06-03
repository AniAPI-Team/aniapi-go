package engine

import "net/http"

// Request is a wrapper to http.Request
type Request struct {
	Data               *http.Request
	Params             []string
	Query              map[string]string
	NeedSingleResource bool
}
