package engine

import (
	"log"
	"net/http"
	"regexp"
	"time"
)

// FHandler is a function type for handler functions
type FHandler func(*Response, *Request)

// Route is a path handled by the router
type Route struct {
	pattern *regexp.Regexp
	handler FHandler
}

// Server is a custom router
type Server struct {
	routes       []Route
	DefaultRoute FHandler
}

func defaultRouteHandler(w *Response, r *Request) {
	w.Writer.Header().Set("Content-Type", "application/json")
	w.Write(http.StatusNotFound, "")
}

func loggingWrapperHandle(handle FHandler) FHandler {
	return func(w *Response, r *Request) {
		start := time.Now()
		handle(w, r)

		elapsed := time.Since(start)
		log.Printf("HTTP %s %s - %d in %dms\n", r.Data.Method, r.Data.URL, w.Status, elapsed.Milliseconds())
	}
}

func setCorsHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, x-fbi-tracking")
	w.Header().Set("Access-Control-Max-Age", "86400")
	w.Header().Set("Vary", " Accept-Encoding, Origin")
	w.Header().Set("Keep-Alive", "timeout=2, max=100 ")
	w.Header().Set("Connection", "Keep-Alive")
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	res := &Response{w, 200, false}
	req := &Request{Data: r}

	setCorsHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(204)
		return
	}

	for _, rt := range s.routes {
		if matches := rt.pattern.FindStringSubmatch(r.URL.Path); len(matches) > 0 {
			if len(matches) > 1 {
				req.Params = matches[1:]
			}

			rt.handler(res, req)

			if res.DefaultError {
				defaultRouteHandler(res, req)
			}

			return
		}
	}

	s.DefaultRoute(res, req)
}

// Handle adds a new route to the router
// All routes are evaluated using insertion order
func (s *Server) Handle(pattern string, handler FHandler) {
	re := regexp.MustCompile("^" + pattern)
	route := Route{pattern: re, handler: loggingWrapperHandle(handler)}

	s.routes = append(s.routes, route)
}

// NewServer creates a new application router
func NewServer() *Server {
	return &Server{
		DefaultRoute: loggingWrapperHandle(defaultRouteHandler),
	}
}
