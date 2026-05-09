package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/javier/go-http-server/internal/handler"
)

func TestHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.Health(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp handler.Envelope[handler.HealthResponse]
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if resp.Data.Status != "ok" {
		t.Errorf("status = %q, want ok", resp.Data.Status)
	}
}

func TestReady(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	handler.Ready(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestItemsCRUD(t *testing.T) {
	mux := handler.NewRouter()

	// Create
	body := bytes.NewBufferString(`{"name":"test item"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/items", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("create: status = %d, want %d\nbody: %s", rec.Code, http.StatusCreated, rec.Body)
	}

	var created handler.Envelope[handler.Item]
	if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
		t.Fatalf("decode created: %v", err)
	}
	if created.Data.Name != "test item" {
		t.Errorf("name = %q, want test item", created.Data.Name)
	}

	// List
	req = httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("list: status = %d, want %d", rec.Code, http.StatusOK)
	}

	// Get by ID
	idStr := strconv.Itoa(created.Data.ID)
	req = httptest.NewRequest(http.MethodGet, "/api/v1/items/"+idStr, nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("get: status = %d, want %d", rec.Code, http.StatusOK)
	}

	// Delete
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/items/"+idStr, nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete: status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	// Get after delete → 404
	req = httptest.NewRequest(http.MethodGet, "/api/v1/items/"+idStr, nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("get deleted: status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestCreateItem_MissingName(t *testing.T) {
	mux := handler.NewRouter()
	body := bytes.NewBufferString(`{"name":""}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/items", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnprocessableEntity)
	}
}

func TestGetItem_InvalidID(t *testing.T) {
	mux := handler.NewRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/items/abc", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
