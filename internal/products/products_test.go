package products_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/javier/go-http-server/internal/platform/config"
	"github.com/javier/go-http-server/internal/platform/database"
	"github.com/javier/go-http-server/internal/platform/server/routes"
	"github.com/javier/go-http-server/internal/products"
)

// newTestRouter builds a router backed by the PostgreSQL repository. These are
// integration tests: if no database is reachable (e.g. `make db-up` was not
// run), they are skipped so `go test ./...` stays green without a database.
//
// Point them at a database with DATABASE_URL; the default targets the local
// docker-compose Postgres.
func newTestRouter(t *testing.T) http.Handler {
	t.Helper()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	pool, err := database.NewPool(context.Background(), cfg)
	if err != nil {
		t.Skipf("skipping integration test, no database available: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := database.Migrate(context.Background(), pool); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	repo := products.NewPostgresRepository(pool)
	if err := repo.Seed(context.Background()); err != nil {
		t.Fatalf("seed: %v", err)
	}

	return routes.NewRouter(repo)
}

// ---- products ---------------------------------------------------------

func TestListProducts(t *testing.T) {
	mux := newTestRouter(t)
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
	mux := newTestRouter(t)
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
	mux := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/999999", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestGetProduct_InvalidID(t *testing.T) {
	mux := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/abc", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestCreateProduct(t *testing.T) {
	mux := newTestRouter(t)
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
	mux := newTestRouter(t)
	body := bytes.NewBufferString(`{"product":{"title":""}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/products", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnprocessableEntity)
	}
}
