package products

import (
	"context"
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
//
// Se devuelve como puntero (*ProductService) por tres razones:
//  1. Sus métodos tienen receptor *ProductService: en Go, si cualquier método
//     usa *T, el valor plano T no satisface la interfaz implícita que espera
//     quien lo consume (ProductController), por lo que hay que devolver *T.
//  2. Evita copiar el struct (con su campo repo) en cada asignación.
//  3. Señala que es una instancia larga, creada una vez y compartida durante
//     toda la vida del proceso, no un dato temporal.
type ProductService struct {
	repo ProductRepository
}

func NewProductService(repo ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) List(ctx context.Context, params ListProductsParams) ([]Product, error) {
	limit := params.Limit
	if limit < 1 || limit > 250 {
		limit = 50
	}
	page := params.Page
	if page < 1 {
		page = 1
	}

	products, err := s.repo.List(ctx, ProductFilters{
		Vendor:      strings.ToLower(params.Vendor),
		ProductType: strings.ToLower(params.ProductType),
		Status:      params.Status,
		Handle:      params.Handle,
	})
	if err != nil {
		return nil, err
	}

	start := (page - 1) * limit
	if start >= len(products) {
		return []Product{}, nil
	}
	end := start + limit
	if end > len(products) {
		end = len(products)
	}
	return products[start:end], nil
}

func (s *ProductService) Count(ctx context.Context) (int, error) {
	return s.repo.Count(ctx)
}

func (s *ProductService) GetByID(ctx context.Context, id int64) (Product, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ProductService) Create(ctx context.Context, input CreateProductInput) (Product, error) {
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

	return s.repo.Create(ctx, product)
}

func (s *ProductService) Update(ctx context.Context, id int64, input UpdateProductInput) (Product, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return Product{}, err
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

	return s.repo.Update(ctx, product)
}

func (s *ProductService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
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
