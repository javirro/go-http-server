package handler

import "time"

// ProductStatus mirrors Shopify's product status values.
type ProductStatus string

const (
	ProductStatusActive   ProductStatus = "active"
	ProductStatusDraft    ProductStatus = "draft"
	ProductStatusArchived ProductStatus = "archived"
)

// FinancialStatus mirrors Shopify's order financial status.
type FinancialStatus string

const (
	FinancialStatusPending    FinancialStatus = "pending"
	FinancialStatusPaid       FinancialStatus = "paid"
	FinancialStatusRefunded   FinancialStatus = "refunded"
	FinancialStatusVoided     FinancialStatus = "voided"
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

// Collection represents a group of products (e.g. "La Liga", "Equipaciones Locales").
type Collection struct {
	ID          int64         `json:"id"`
	Title       string        `json:"title"`
	Handle      string        `json:"handle"`
	BodyHTML    string        `json:"body_html"`
	Image       *ProductImage `json:"image"`
	ProductIDs  []int64       `json:"product_ids,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	PublishedAt *time.Time    `json:"published_at"`
}

// Cart represents a shopping cart in progress.
type Cart struct {
	Token      string     `json:"token"`
	ItemCount  int        `json:"item_count"`
	TotalPrice string     `json:"total_price"`
	Currency   string     `json:"currency"`
	Items      []CartItem `json:"items"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// CartItem is a line item inside a cart.
type CartItem struct {
	ID           int64   `json:"id"`
	VariantID    int64   `json:"variant_id"`
	ProductID    int64   `json:"product_id"`
	Title        string  `json:"title"`
	VariantTitle string  `json:"variant_title"`
	SKU          string  `json:"sku"`
	Quantity     int     `json:"quantity"`
	Price        string  `json:"price"`
	LinePrice    string  `json:"line_price"`
	Image        *string `json:"image"`
}

// Order represents a completed purchase.
type Order struct {
	ID                int64           `json:"id"`
	OrderNumber       int             `json:"order_number"`
	Email             string          `json:"email"`
	Currency          string          `json:"currency"`
	FinancialStatus   FinancialStatus `json:"financial_status"`
	FulfillmentStatus *string         `json:"fulfillment_status"`
	SubtotalPrice     string          `json:"subtotal_price"`
	TotalTax          string          `json:"total_tax"`
	TotalPrice        string          `json:"total_price"`
	LineItems         []OrderLineItem `json:"line_items"`
	ShippingAddress   *Address        `json:"shipping_address"`
	BillingAddress    *Address        `json:"billing_address"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	ProcessedAt       time.Time       `json:"processed_at"`
}

// OrderLineItem is a product line inside an order.
type OrderLineItem struct {
	ID                int64   `json:"id"`
	VariantID         int64   `json:"variant_id"`
	ProductID         int64   `json:"product_id"`
	Title             string  `json:"title"`
	VariantTitle      string  `json:"variant_title"`
	SKU               string  `json:"sku"`
	Quantity          int     `json:"quantity"`
	Price             string  `json:"price"`
	TotalDiscount     string  `json:"total_discount"`
	Taxable           bool    `json:"taxable"`
	FulfillmentStatus *string `json:"fulfillment_status"`
}

// Address holds shipping or billing address information.
type Address struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Address1  string `json:"address1"`
	Address2  string `json:"address2,omitempty"`
	City      string `json:"city"`
	Province  string `json:"province,omitempty"`
	Country   string `json:"country"`
	Zip       string `json:"zip"`
	Phone     string `json:"phone,omitempty"`
	Company   string `json:"company,omitempty"`
}

// --- Shopify-style request/response wrappers ---

type ProductsResponse struct {
	Products []Product `json:"products"`
}

type ProductResponse struct {
	Product Product `json:"product"`
}

type CollectionsResponse struct {
	Collections []Collection `json:"collections"`
}

type CollectionResponse struct {
	Collection Collection `json:"collection"`
}

type CountResponse struct {
	Count int `json:"count"`
}

type OrdersResponse struct {
	Orders []Order `json:"orders"`
}

type OrderResponse struct {
	Order Order `json:"order"`
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
		Title       string        `json:"title"`
		BodyHTML    string        `json:"body_html"`
		Vendor      string        `json:"vendor"`
		Tags        string        `json:"tags"`
		Status      ProductStatus `json:"status"`
	} `json:"product"`
}

type AddCartItemInput struct {
	ID       int64 `json:"id"`       // variant ID
	Quantity int   `json:"quantity"`
}

type UpdateCartItemInput struct {
	Quantity int `json:"quantity"`
}

type CreateOrderInput struct {
	Order struct {
		CartToken       string   `json:"cart_token"`
		Email           string   `json:"email"`
		ShippingAddress *Address `json:"shipping_address"`
		BillingAddress  *Address `json:"billing_address"`
	} `json:"order"`
}
