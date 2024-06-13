package http

import "net/http"

// Router enables connecting a REST API to the echo server. The API wrappers should implement this interface
type Router interface {
	// Routes configures the HTTP routes on the given router
	Routes(router *http.ServeMux)
}
