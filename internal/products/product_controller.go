package products

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/javier/go-http-server/internal/platform/server/respond"
)

// ProductController maps HTTP requests to product service operations.
type ProductController struct {
	service *ProductService
}

func NewProductController(service *ProductService) *ProductController {
	return &ProductController{service: service}
}

// List handles GET /api/v1/products
func (c *ProductController) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	products := c.service.List(ListProductsParams{
		Limit:       queryInt(q.Get("limit"), 50),
		Page:        queryInt(q.Get("page"), 1),
		Vendor:      q.Get("vendor"),
		ProductType: q.Get("product_type"),
		Status:      q.Get("status"),
		Handle:      q.Get("handle"),
	})
	respond.JSON(w, r, http.StatusOK, ProductsResponse{Products: products})
}

// Count handles GET /api/v1/products/count
func (c *ProductController) Count(w http.ResponseWriter, r *http.Request) {
	respond.JSON(w, r, http.StatusOK, CountResponse{Count: c.service.Count()})
}

// Get handles GET /api/v1/products/{id}
func (c *ProductController) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	product, err := c.service.GetByID(id)
	if err != nil {
		c.writeServiceError(w, r, err)
		return
	}

	respond.JSON(w, r, http.StatusOK, ProductResponse{Product: product})
}

// Create handles POST /api/v1/products
func (c *ProductController) Create(w http.ResponseWriter, r *http.Request) {
	var input CreateProductInput
	if !respond.DecodeJSON(w, r, &input) {
		return
	}

	product, err := c.service.Create(input)
	if err != nil {
		c.writeServiceError(w, r, err)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/api/v1/products/%d", product.ID))
	respond.JSON(w, r, http.StatusCreated, ProductResponse{Product: product})
}

// Update handles PUT /api/v1/products/{id}
func (c *ProductController) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	var input UpdateProductInput
	if !respond.DecodeJSON(w, r, &input) {
		return
	}

	product, err := c.service.Update(id, input)
	if err != nil {
		c.writeServiceError(w, r, err)
		return
	}

	respond.JSON(w, r, http.StatusOK, ProductResponse{Product: product})
}

// Delete handles DELETE /api/v1/products/{id}
func (c *ProductController) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	if err := c.service.Delete(id); err != nil {
		c.writeServiceError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *ProductController) writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	var vErr ValidationError
	switch {
	case errors.Is(err, ErrProductNotFound):
		respond.ErrorResponse(w, r, http.StatusNotFound, "product not found")
	case errors.As(err, &vErr):
		respond.ErrorResponse(w, r, http.StatusUnprocessableEntity, vErr.Message)
	default:
		respond.ErrorResponse(w, r, http.StatusInternalServerError, "internal server error")
	}
}

func parseID(w http.ResponseWriter, r *http.Request, param string) (int64, bool) {
	raw := r.PathValue(param)
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id < 1 {
		respond.ErrorResponse(w, r, http.StatusBadRequest, fmt.Sprintf("invalid %s", param))
		return 0, false
	}
	return id, true
}

func queryInt(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
