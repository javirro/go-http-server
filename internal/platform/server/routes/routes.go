package routes

import (
	"net/http"

	"github.com/javier/go-http-server/internal/products"
)

// NewRouter builds the application's HTTP router and wires every feature's
// routes onto it, using the provided dependencies. Go 1.22+ pattern syntax is
// used: "METHOD /path/{param}".
func NewRouter(productRepo products.ProductRepository) *http.ServeMux {
	mux := http.NewServeMux()

	products.Register(mux, productRepo)

	return mux
}
