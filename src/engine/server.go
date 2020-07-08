package engine

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/gorilla/websocket"
)

// FHandler is a function type for handler functions
type FHandler func(*Response, *Request)

// Route is a path handled by the router
type Route struct {
	pattern *regexp.Regexp
	handler FHandler
}

// SocketMessage is the standard socket message type
type SocketMessage struct {
	Channel string      `json:"channel"`
	Data    interface{} `json:"data"`
}

// Server is a custom router
type Server struct {
	routes       []Route
	DefaultRoute FHandler
}

var socketConnections []*websocket.Conn
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     checkOrigin,
}

func checkOrigin(r *http.Request) bool {
	return true
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

// OnSocketConnStart adds a new socket connection to the alive connections pool
func OnSocketConnStart(w *Response, r *Request) {
	conn, err := upgrader.Upgrade(w.Writer, r.Data, nil)

	if err != nil {
		log.Printf("SOCKET ERROR: %s", err.Error())
		return
	}

	duplicate := false

	for _, c := range socketConnections {
		if conn.RemoteAddr().String() == c.RemoteAddr().String() {
			duplicate = true
		}
	}

	if duplicate == false {
		socketConnections = append(socketConnections, conn)
	}
}

// SocketWriteMessage writes a message to all alive connections
func SocketWriteMessage(msg *SocketMessage) {
	var toDelete []int

	json, err := json.Marshal(msg)

	if err != nil {
		log.Printf("SOCKET JSON ERROR: %s", err.Error())
		return
	}

	for i, s := range socketConnections {
		err := s.WriteMessage(1, []byte(json))

		if err != nil {
			log.Printf("SOCKET WRITE ERROR: %s", err.Error())
			toDelete = append(toDelete, i)
		}
	}

	for j := len(toDelete) - 1; j >= 0; j-- {
		i := toDelete[j]

		length := len(socketConnections) - 1
		socketConnections[i] = socketConnections[length]
		socketConnections[length] = nil
		socketConnections = socketConnections[:length]
	}
}

// NewServer creates a new application router
func NewServer() *Server {
	return &Server{
		DefaultRoute: loggingWrapperHandle(defaultRouteHandler),
	}
}
