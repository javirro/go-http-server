package products

import "time"

// ProductStatus mirrors Shopify's product status values.
type ProductStatus string

const (
	ProductStatusActive   ProductStatus = "active"
	ProductStatusDraft    ProductStatus = "draft"
	ProductStatusArchived ProductStatus = "archived"
)

// Product represents a Shopify-like product (a football team jersey).
type Product struct {
	ID          int64           `json:"id"`
	Title       string          `json:"title"`
	Handle      string          `json:"handle"`
	BodyHTML    string          `json:"body_html"`
	Vendor      string          `json:"vendor"`
	ProductType string          `json:"product_type"`
	Tags        string          `json:"tags"`
	Status      ProductStatus   `json:"status"`
	Options     []ProductOption `json:"options"`
	Variants    []Variant       `json:"variants"`
	Images      []ProductImage  `json:"images"`
	Image       *ProductImage   `json:"image"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	PublishedAt *time.Time      `json:"published_at"`
}

// ProductOption represents a product option (e.g. "Talla").
type ProductOption struct {
	ID        int64    `json:"id"`
	ProductID int64    `json:"product_id"`
	Name      string   `json:"name"`
	Position  int      `json:"position"`
	Values    []string `json:"values"`
}

// Variant represents a specific size/colour variant of a product.
type Variant struct {
	ID                  int64     `json:"id"`
	ProductID           int64     `json:"product_id"`
	Title               string    `json:"title"`
	SKU                 string    `json:"sku"`
	Position            int       `json:"position"`
	Price               string    `json:"price"`
	CompareAtPrice      *string   `json:"compare_at_price"`
	InventoryQuantity   int       `json:"inventory_quantity"`
	Option1             string    `json:"option1"`
	Option2             *string   `json:"option2"`
	Option3             *string   `json:"option3"`
	Weight              float64   `json:"weight"`
	WeightUnit          string    `json:"weight_unit"`
	RequiresShipping    bool      `json:"requires_shipping"`
	Taxable             bool      `json:"taxable"`
	FulfillmentService  string    `json:"fulfillment_service"`
	InventoryManagement string    `json:"inventory_management"`
	InventoryPolicy     string    `json:"inventory_policy"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// ProductImage represents a product image.
type ProductImage struct {
	ID         int64     `json:"id"`
	ProductID  int64     `json:"product_id"`
	Position   int       `json:"position"`
	Src        string    `json:"src"`
	Width      int       `json:"width"`
	Height     int       `json:"height"`
	Alt        *string   `json:"alt"`
	VariantIDs []int64   `json:"variant_ids"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// --- Shopify-style response wrappers ---

type ProductsResponse struct {
	Products []Product `json:"products"`
}

type ProductResponse struct {
	Product Product `json:"product"`
}

type CountResponse struct {
	Count int `json:"count"`
}

// --- Request bodies ---

type CreateProductInput struct {
	Product struct {
		Title       string        `json:"title"`
		Handle      string        `json:"handle"`
		BodyHTML    string        `json:"body_html"`
		Vendor      string        `json:"vendor"`
		ProductType string        `json:"product_type"`
		Tags        string        `json:"tags"`
		Status      ProductStatus `json:"status"`
	} `json:"product"`
}

type UpdateProductInput struct {
	Product struct {
		Title    string        `json:"title"`
		BodyHTML string        `json:"body_html"`
		Vendor   string        `json:"vendor"`
		Tags     string        `json:"tags"`
		Status   ProductStatus `json:"status"`
	} `json:"product"`
}
