package products

import (
	"errors"
	"strings"
	"time"
)

var ErrProductNotFound = errors.New("product not found")

type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string { return e.Message }

// ListProductsParams defines service-level query options.
type ListProductsParams struct {
	Limit       int
	Page        int
	Vendor      string
	ProductType string
	Status      string
	Handle      string
}

// ProductService holds product business logic.
type ProductService struct {
	repo ProductRepository
}

func NewProductService(repo ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) List(params ListProductsParams) []Product {
	limit := params.Limit
	if limit < 1 || limit > 250 {
		limit = 50
	}
	page := params.Page
	if page < 1 {
		page = 1
	}

	products := s.repo.List(ProductFilters{
		Vendor:      strings.ToLower(params.Vendor),
		ProductType: strings.ToLower(params.ProductType),
		Status:      params.Status,
		Handle:      params.Handle,
	})

	start := (page - 1) * limit
	if start >= len(products) {
		return []Product{}
	}
	end := start + limit
	if end > len(products) {
		end = len(products)
	}
	return products[start:end]
}

func (s *ProductService) Count() int {
	return s.repo.Count()
}

func (s *ProductService) GetByID(id int64) (Product, error) {
	product, exists := s.repo.GetByID(id)
	if !exists {
		return Product{}, ErrProductNotFound
	}
	return product, nil
}

func (s *ProductService) Create(input CreateProductInput) (Product, error) {
	title := strings.TrimSpace(input.Product.Title)
	if title == "" {
		return Product{}, ValidationError{Message: "product title is required"}
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
		handle = slugify(title)
	}

	product := Product{
		Title:       title,
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

	return s.repo.Create(product), nil
}

func (s *ProductService) Update(id int64, input UpdateProductInput) (Product, error) {
	product, exists := s.repo.GetByID(id)
	if !exists {
		return Product{}, ErrProductNotFound
	}

	if t := strings.TrimSpace(input.Product.Title); t != "" {
		product.Title = t
	}
	if input.Product.BodyHTML != "" {
		product.BodyHTML = input.Product.BodyHTML
	}
	if input.Product.Vendor != "" {
		product.Vendor = input.Product.Vendor
	}
	if input.Product.Tags != "" {
		product.Tags = input.Product.Tags
	}
	if input.Product.Status != "" {
		product.Status = input.Product.Status
		if product.Status == ProductStatusActive && product.PublishedAt == nil {
			t := time.Now().UTC()
			product.PublishedAt = &t
		}
	}
	product.UpdatedAt = time.Now().UTC()

	updated, ok := s.repo.Update(product)
	if !ok {
		return Product{}, ErrProductNotFound
	}
	return updated, nil
}

func (s *ProductService) Delete(id int64) error {
	if ok := s.repo.Delete(id); !ok {
		return ErrProductNotFound
	}
	return nil
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
