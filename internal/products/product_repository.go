package products

import "strings"

// ProductFilters defines repository-level filters for product listing.
type ProductFilters struct {
	Vendor      string
	ProductType string
	Status      string
	Handle      string
}

// ProductRepository defines persistence operations for products.
type ProductRepository interface {
	List(filters ProductFilters) []Product
	Count() int
	GetByID(id int64) (Product, bool)
	Create(product Product) Product
	Update(product Product) (Product, bool)
	Delete(id int64) bool
}

// InMemoryProductRepository persists products in the shared in-memory store.
type InMemoryProductRepository struct {
	store *shopStore
}

func NewInMemoryProductRepository(store *shopStore) *InMemoryProductRepository {
	return &InMemoryProductRepository{store: store}
}

func (r *InMemoryProductRepository) List(filters ProductFilters) []Product {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()

	var products []Product
	for _, p := range r.store.products {
		if filters.Vendor != "" && strings.ToLower(p.Vendor) != filters.Vendor {
			continue
		}
		if filters.ProductType != "" && strings.ToLower(p.ProductType) != filters.ProductType {
			continue
		}
		if filters.Status != "" && string(p.Status) != filters.Status {
			continue
		}
		if filters.Handle != "" && p.Handle != filters.Handle {
			continue
		}
		products = append(products, *p)
	}

	// Deterministic ordering by ID.
	sortProductsByID(products)
	return products
}

func (r *InMemoryProductRepository) Count() int {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return len(r.store.products)
}

func (r *InMemoryProductRepository) GetByID(id int64) (Product, bool) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()

	p, exists := r.store.products[id]
	if !exists {
		return Product{}, false
	}
	return *p, true
}

func (r *InMemoryProductRepository) Create(product Product) Product {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	id := r.store.nextProductID()
	product.ID = id
	p := product
	r.store.products[id] = &p
	return p
}

func (r *InMemoryProductRepository) Update(product Product) (Product, bool) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	if _, exists := r.store.products[product.ID]; !exists {
		return Product{}, false
	}
	p := product
	r.store.products[product.ID] = &p
	return p, true
}

func (r *InMemoryProductRepository) Delete(id int64) bool {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	if _, exists := r.store.products[id]; !exists {
		return false
	}
	delete(r.store.products, id)
	return true
}

func sortProductsByID(ps []Product) {
	for i := 1; i < len(ps); i++ {
		for j := i; j > 0 && ps[j].ID < ps[j-1].ID; j-- {
			ps[j], ps[j-1] = ps[j-1], ps[j]
		}
	}
}
