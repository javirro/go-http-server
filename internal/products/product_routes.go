package products

import "net/http"

// Register wires the product feature (service → controller) on top of the
// given repository and registers its HTTP routes on the mux.
func Register(mux *http.ServeMux, repo ProductRepository) {
	service := NewProductService(repo)
	controller := NewProductController(service)
	registerProductRoutes(mux, controller)
}

func registerProductRoutes(mux *http.ServeMux, controller *ProductController) {
	// Products — /api/v1/products
	// Note: the literal /count pattern must be registered before /{id} so that
	// Go's ServeMux gives it priority.
	mux.HandleFunc("GET /api/v1/products/count", controller.Count)
	mux.HandleFunc("GET /api/v1/products", controller.List)
	mux.HandleFunc("POST /api/v1/products", controller.Create)
	mux.HandleFunc("GET /api/v1/products/{id}", controller.Get)
	mux.HandleFunc("PUT /api/v1/products/{id}", controller.Update)
	mux.HandleFunc("DELETE /api/v1/products/{id}", controller.Delete)
}
