package products_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/javier/go-http-server/internal/platform/server/routes"
	"github.com/javier/go-http-server/internal/products"
)

// ---- products ---------------------------------------------------------

func TestListProducts(t *testing.T) {
	mux := routes.NewRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp products.ProductsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(resp.Products) == 0 {
		t.Error("expected seeded products, got none")
	}
}

func TestCountProducts(t *testing.T) {
	mux := routes.NewRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/count", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp products.CountResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if resp.Count == 0 {
		t.Error("expected non-zero product count")
	}
}

func TestGetProduct_NotFound(t *testing.T) {
	mux := routes.NewRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/999999", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestGetProduct_InvalidID(t *testing.T) {
	mux := routes.NewRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/abc", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestCreateProduct(t *testing.T) {
	mux := routes.NewRouter()
	body := bytes.NewBufferString(`{"product":{"title":"Camiseta Test FC","vendor":"Test Brand","status":"draft"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/products", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d\nbody: %s", rec.Code, http.StatusCreated, rec.Body)
	}

	var resp products.ProductResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Product.Title != "Camiseta Test FC" {
		t.Errorf("title = %q, want %q", resp.Product.Title, "Camiseta Test FC")
	}
	if resp.Product.ID == 0 {
		t.Error("expected non-zero product id")
	}
}

func TestCreateProduct_MissingTitle(t *testing.T) {
	mux := routes.NewRouter()
	body := bytes.NewBufferString(`{"product":{"title":""}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/products", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnprocessableEntity)
	}
}
