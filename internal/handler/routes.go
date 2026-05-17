package handler

import (
	"net/http"
)

// NewRouter registers all application routes on a new ServeMux.
// Go 1.22+ pattern syntax is used: "METHOD /path/{param}".
func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()

	// Products — /api/v1/products
	// Note: the literal /count pattern must be registered before /{id} so that
	// Go's ServeMux gives it priority.
	mux.HandleFunc("GET /api/v1/products/count", CountProducts)
	mux.HandleFunc("GET /api/v1/products", ListProducts)
	mux.HandleFunc("POST /api/v1/products", CreateProduct)
	mux.HandleFunc("GET /api/v1/products/{id}", GetProduct)
	mux.HandleFunc("PUT /api/v1/products/{id}", UpdateProduct)
	mux.HandleFunc("DELETE /api/v1/products/{id}", DeleteProduct)

	return mux
}
