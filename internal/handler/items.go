package handler

import (
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Item is the domain model used in the example CRUD API.
type Item struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// itemStore is a simple in-memory store (replace with a real DB in production).
type itemStore struct {
	mu      sync.RWMutex
	items   map[int]Item
	counter int
}

var store = &itemStore{items: make(map[int]Item)}

// ListItems handles GET /api/v1/items
func ListItems(w http.ResponseWriter, r *http.Request) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	items := make([]Item, 0, len(store.items))
	for _, item := range store.items {
		items = append(items, item)
	}

	JSON(w, r, http.StatusOK, Envelope[[]Item]{Data: items})
}

// CreateItem handles POST /api/v1/items
func CreateItem(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string `json:"name"`
	}
	if !DecodeJSON(w, r, &input) {
		return
	}

	if input.Name == "" {
		ErrorResponse(w, r, http.StatusUnprocessableEntity, "name is required")
		return
	}

	store.mu.Lock()
	store.counter++
	item := Item{
		ID:        store.counter,
		Name:      input.Name,
		CreatedAt: time.Now().UTC(),
	}
	store.items[item.ID] = item
	store.mu.Unlock()

	w.Header().Set("Location", "/api/v1/items/"+strconv.Itoa(item.ID))
	JSON(w, r, http.StatusCreated, Envelope[Item]{Data: item})
}

// GetItem handles GET /api/v1/items/{id}
func GetItem(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		ErrorResponse(w, r, http.StatusBadRequest, "invalid item id")
		return
	}

	store.mu.RLock()
	item, ok := store.items[id]
	store.mu.RUnlock()

	if !ok {
		ErrorResponse(w, r, http.StatusNotFound, "item not found")
		return
	}

	JSON(w, r, http.StatusOK, Envelope[Item]{Data: item})
}

// DeleteItem handles DELETE /api/v1/items/{id}
func DeleteItem(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		ErrorResponse(w, r, http.StatusBadRequest, "invalid item id")
		return
	}

	store.mu.Lock()
	_, ok := store.items[id]
	if ok {
		delete(store.items, id)
	}
	store.mu.Unlock()

	if !ok {
		ErrorResponse(w, r, http.StatusNotFound, "item not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
