package handler

import (
	"net/http"
)

// NewRouter registers all application routes on a new ServeMux.
// Go 1.22+ pattern syntax is used: "METHOD /path/{param}".
func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()

	// Observability
	mux.HandleFunc("GET /health", Health)
	mux.HandleFunc("GET /ready", Ready)

	// Items API
	mux.HandleFunc("GET /api/v1/items", ListItems)
	mux.HandleFunc("POST /api/v1/items", CreateItem)
	mux.HandleFunc("GET /api/v1/items/{id}", GetItem)
	mux.HandleFunc("DELETE /api/v1/items/{id}", DeleteItem)

	return mux
}
