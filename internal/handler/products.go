package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ListProducts handles GET /api/v1/products
//
// Supported query params:
//
//	limit        – max number of results (default 50, max 250)
//	page         – 1-based page number (default 1)
//	vendor       – filter by vendor
//	product_type – filter by product type
//	status       – filter by status (active|draft|archived)
//	handle       – filter by handle (exact match)
func ListProducts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	limit := queryInt(q.Get("limit"), 50)
	if limit < 1 || limit > 250 {
		limit = 50
	}
	page := queryInt(q.Get("page"), 1)
	if page < 1 {
		page = 1
	}

	vendor := strings.ToLower(q.Get("vendor"))
	productType := strings.ToLower(q.Get("product_type"))
	status := ProductStatus(q.Get("status"))
	handle := q.Get("handle")

	shop.mu.RLock()
	defer shop.mu.RUnlock()

	var filtered []Product
	for _, p := range shop.products {
		if vendor != "" && strings.ToLower(p.Vendor) != vendor {
			continue
		}
		if productType != "" && strings.ToLower(p.ProductType) != productType {
			continue
		}
		if status != "" && p.Status != status {
			continue
		}
		if handle != "" && p.Handle != handle {
			continue
		}
		filtered = append(filtered, *p)
	}

	// Deterministic ordering by ID.
	sortProductsByID(filtered)

	start := (page - 1) * limit
	if start >= len(filtered) {
		JSON(w, r, http.StatusOK, ProductsResponse{Products: []Product{}})
		return
	}
	end := start + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	JSON(w, r, http.StatusOK, ProductsResponse{Products: filtered[start:end]})
}

// CountProducts handles GET /api/v1/products/count
func CountProducts(w http.ResponseWriter, r *http.Request) {
	shop.mu.RLock()
	n := len(shop.products)
	shop.mu.RUnlock()

	JSON(w, r, http.StatusOK, CountResponse{Count: n})
}

// GetProduct handles GET /api/v1/products/{id}
func GetProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	shop.mu.RLock()
	p, exists := shop.products[id]
	shop.mu.RUnlock()

	if !exists {
		ErrorResponse(w, r, http.StatusNotFound, "product not found")
		return
	}

	JSON(w, r, http.StatusOK, ProductResponse{Product: *p})
}

// CreateProduct handles POST /api/v1/products
func CreateProduct(w http.ResponseWriter, r *http.Request) {
	var input CreateProductInput
	if !DecodeJSON(w, r, &input) {
		return
	}

	if strings.TrimSpace(input.Product.Title) == "" {
		ErrorResponse(w, r, http.StatusUnprocessableEntity, "product title is required")
		return
	}

	status := input.Product.Status
	if status == "" {
		status = ProductStatusDraft
	}

	now := time.Now().UTC()
	var publishedAt *time.Time
	if status == ProductStatusActive {
		t := now
		publishedAt = &t
	}

	handle := input.Product.Handle
	if handle == "" {
		handle = slugify(input.Product.Title)
	}

	shop.mu.Lock()
	id := shop.nextProductID()
	p := &Product{
		ID:          id,
		Title:       input.Product.Title,
		Handle:      handle,
		BodyHTML:    input.Product.BodyHTML,
		Vendor:      input.Product.Vendor,
		ProductType: input.Product.ProductType,
		Tags:        input.Product.Tags,
		Status:      status,
		Options:     []ProductOption{},
		Variants:    []Variant{},
		Images:      []ProductImage{},
		CreatedAt:   now,
		UpdatedAt:   now,
		PublishedAt: publishedAt,
	}
	shop.products[id] = p
	shop.mu.Unlock()

	w.Header().Set("Location", fmt.Sprintf("/api/v1/products/%d", id))
	JSON(w, r, http.StatusCreated, ProductResponse{Product: *p})
}

// UpdateProduct handles PUT /api/v1/products/{id}
func UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	var input UpdateProductInput
	if !DecodeJSON(w, r, &input) {
		return
	}

	shop.mu.Lock()
	p, exists := shop.products[id]
	if !exists {
		shop.mu.Unlock()
		ErrorResponse(w, r, http.StatusNotFound, "product not found")
		return
	}

	if t := strings.TrimSpace(input.Product.Title); t != "" {
		p.Title = t
	}
	if input.Product.BodyHTML != "" {
		p.BodyHTML = input.Product.BodyHTML
	}
	if input.Product.Vendor != "" {
		p.Vendor = input.Product.Vendor
	}
	if input.Product.Tags != "" {
		p.Tags = input.Product.Tags
	}
	if input.Product.Status != "" {
		p.Status = input.Product.Status
		if p.Status == ProductStatusActive && p.PublishedAt == nil {
			t := time.Now().UTC()
			p.PublishedAt = &t
		}
	}
	p.UpdatedAt = time.Now().UTC()

	updated := *p
	shop.mu.Unlock()

	JSON(w, r, http.StatusOK, ProductResponse{Product: updated})
}

// DeleteProduct handles DELETE /api/v1/products/{id}
func DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	shop.mu.Lock()
	_, exists := shop.products[id]
	if exists {
		delete(shop.products, id)
	}
	shop.mu.Unlock()

	if !exists {
		ErrorResponse(w, r, http.StatusNotFound, "product not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- helpers ---

func parseID(w http.ResponseWriter, r *http.Request, param string) (int64, bool) {
	raw := r.PathValue(param)
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id < 1 {
		ErrorResponse(w, r, http.StatusBadRequest, fmt.Sprintf("invalid %s", param))
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

func slugify(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ' || r == '_':
			b.WriteRune('-')
		case r == 'á' || r == 'à' || r == 'ä' || r == 'â':
			b.WriteRune('a')
		case r == 'é' || r == 'è' || r == 'ë' || r == 'ê':
			b.WriteRune('e')
		case r == 'í' || r == 'ì' || r == 'ï' || r == 'î':
			b.WriteRune('i')
		case r == 'ó' || r == 'ò' || r == 'ö' || r == 'ô':
			b.WriteRune('o')
		case r == 'ú' || r == 'ù' || r == 'ü' || r == 'û':
			b.WriteRune('u')
		case r == 'ñ':
			b.WriteRune('n')
		}
	}
	return b.String()
}

func sortProductsByID(ps []Product) {
	for i := 1; i < len(ps); i++ {
		for j := i; j > 0 && ps[j].ID < ps[j-1].ID; j-- {
			ps[j], ps[j-1] = ps[j-1], ps[j]
		}
	}
}
