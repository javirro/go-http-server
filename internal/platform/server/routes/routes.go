package routes

import (
	"net/http"

	"github.com/javier/go-http-server/internal/products"
)

// NewRouter builds the application's HTTP router and wires every feature's
// routes onto it. Go 1.22+ pattern syntax is used: "METHOD /path/{param}".
func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()

	products.Register(mux)

	return mux
}
