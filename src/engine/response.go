package engine

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

// Response is a wrapper to http.Response
type Response struct {
	Writer       http.ResponseWriter
	Status       int
	DefaultError bool
}

func (res *Response) Write(status int, body string) {
	res.Status = status
	res.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	res.Writer.WriteHeader(status)
	fmt.Fprint(res.Writer, body)
}

// NotFound is used to setup the response to use default route handler
func (res *Response) NotFound() {
	res.DefaultError = true
	res.Status = http.StatusNotFound
}

// NotImplemented is used to setup the response to 501 status code
func (res *Response) NotImplemented() {
	res.DefaultError = false
	res.Write(http.StatusNotImplemented, "Not implemented")
}

// TooManyRequests is used to setup the response to 429 status code
func (res *Response) TooManyRequests() {
	res.DefaultError = false
	res.Write(http.StatusTooManyRequests, "Too many requests")
}

// NotAuthorized is used to setup the response to 401 status code
func (res *Response) NotAuthorized() {
	res.DefaultError = false
	res.Write(http.StatusUnauthorized, "Not authorized")
}

// BadRequest is used to setup the response to 400 status code
func (res *Response) BadRequest() {
	res.DefaultError = false
	res.Write(http.StatusBadRequest, "Bad Request")
}

// WriteJSON is used to setup the response to use JSON format
func (res *Response) WriteJSON(status int, body string) {
	res.Writer.Header().Set("Content-Type", "application/json")
	res.Write(status, body)
}

// WriteJSONError is used to setup the response error to use JSON format
func (res *Response) WriteJSONError(status int, body string) {
	res.Writer.Header().Set("Content-Type", "application/json")
	res.Write(status, "{ \"error\": \""+body+"\"}")
}

// WriteHTML is used to setup an HTML file response content
func (res *Response) WriteHTML(filePath string) {
	html, err := ioutil.ReadFile(filePath)

	if err != nil {
		res.NotFound()
		return
	}

	res.Write(200, string(html))
}
