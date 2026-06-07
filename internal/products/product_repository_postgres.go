package products

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProductFilters defines repository-level filters for product listing.
type ProductFilters struct {
	Vendor      string
	ProductType string
	Status      string
	Handle      string
}

// ProductRepository defines persistence operations for products. Methods take a
// context and return errors so implementations can talk to a real database.
// A missing product is reported with ErrProductNotFound.
type ProductRepository interface {
	List(ctx context.Context, filters ProductFilters) ([]Product, error)
	Count(ctx context.Context) (int, error)
	GetByID(ctx context.Context, id int64) (Product, error)
	Create(ctx context.Context, product Product) (Product, error)
	Update(ctx context.Context, product Product) (Product, error)
	Delete(ctx context.Context, id int64) error
}

// productColumns lists the columns selected/returned for a product row, in the
// order expected by scanProduct.
const productColumns = `id, title, handle, body_html, vendor, product_type, tags, status,
	options, variants, images, created_at, updated_at, published_at`

// PostgresProductRepository persists products in PostgreSQL using pgxpool.
// Nested collections (options, variants, images) are stored as JSONB.
type PostgresProductRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresProductRepository {
	return &PostgresProductRepository{pool: pool}
}

// rowScanner is satisfied by both pgx.Row and pgx.Rows.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanProduct(row rowScanner) (Product, error) {
	var p Product
	var status string

	if err := row.Scan(
		&p.ID, &p.Title, &p.Handle, &p.BodyHTML, &p.Vendor, &p.ProductType, &p.Tags,
		&status, &p.Options, &p.Variants, &p.Images, &p.CreatedAt, &p.UpdatedAt, &p.PublishedAt,
	); err != nil {
		return Product{}, err
	}

	p.Status = ProductStatus(status)
	// The model exposes a convenience "image" (the first one) alongside "images".
	if len(p.Images) > 0 {
		img := p.Images[0]
		p.Image = &img
	}
	return p, nil
}

func (r *PostgresProductRepository) List(ctx context.Context, filters ProductFilters) ([]Product, error) {
	query := `SELECT ` + productColumns + ` FROM products`

	var conds []string
	var args []any
	i := 1

	if filters.Vendor != "" {
		conds = append(conds, fmt.Sprintf("LOWER(vendor) = $%d", i))
		args = append(args, filters.Vendor)
		i++
	}
	if filters.ProductType != "" {
		conds = append(conds, fmt.Sprintf("LOWER(product_type) = $%d", i))
		args = append(args, filters.ProductType)
		i++
	}
	if filters.Status != "" {
		conds = append(conds, fmt.Sprintf("status = $%d", i))
		args = append(args, filters.Status)
		i++
	}
	if filters.Handle != "" {
		conds = append(conds, fmt.Sprintf("handle = $%d", i))
		args = append(args, filters.Handle)
		i++
	}

	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}
	query += " ORDER BY id"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query products: %w", err)
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate products: %w", err)
	}
	return products, nil
}

func (r *PostgresProductRepository) Count(ctx context.Context) (int, error) {
	var n int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM products`).Scan(&n); err != nil {
		return 0, fmt.Errorf("count products: %w", err)
	}
	return n, nil
}

func (r *PostgresProductRepository) GetByID(ctx context.Context, id int64) (Product, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+productColumns+` FROM products WHERE id = $1`, id)
	p, err := scanProduct(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Product{}, ErrProductNotFound
	}
	if err != nil {
		return Product{}, fmt.Errorf("get product: %w", err)
	}
	return p, nil
}

func (r *PostgresProductRepository) Create(ctx context.Context, product Product) (Product, error) {
	const q = `INSERT INTO products
		(title, handle, body_html, vendor, product_type, tags, status, options, variants, images, created_at, updated_at, published_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id`

	if err := r.pool.QueryRow(ctx, q,
		product.Title, product.Handle, product.BodyHTML, product.Vendor, product.ProductType,
		product.Tags, string(product.Status), product.Options, product.Variants, product.Images,
		product.CreatedAt, product.UpdatedAt, product.PublishedAt,
	).Scan(&product.ID); err != nil {
		return Product{}, fmt.Errorf("insert product: %w", err)
	}
	return product, nil
}

func (r *PostgresProductRepository) Update(ctx context.Context, product Product) (Product, error) {
	const q = `UPDATE products SET
		title = $2, handle = $3, body_html = $4, vendor = $5, product_type = $6, tags = $7,
		status = $8, options = $9, variants = $10, images = $11, updated_at = $12, published_at = $13
		WHERE id = $1`

	tag, err := r.pool.Exec(ctx, q,
		product.ID, product.Title, product.Handle, product.BodyHTML, product.Vendor,
		product.ProductType, product.Tags, string(product.Status), product.Options,
		product.Variants, product.Images, product.UpdatedAt, product.PublishedAt,
	)
	if err != nil {
		return Product{}, fmt.Errorf("update product: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return Product{}, ErrProductNotFound
	}
	return product, nil
}

func (r *PostgresProductRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete product: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrProductNotFound
	}
	return nil
}

// Seed inserts the catalogue of seeded products if the table is empty. It is a
// no-op once products exist, so it is safe to call on every startup.
func (r *PostgresProductRepository) Seed(ctx context.Context) error {
	var count int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM products`).Scan(&count); err != nil {
		return fmt.Errorf("count products: %w", err)
	}
	if count > 0 {
		return nil
	}

	const q = `INSERT INTO products
		(id, title, handle, body_html, vendor, product_type, tags, status, options, variants, images, created_at, updated_at, published_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	for _, p := range seedProducts() {
		if _, err := r.pool.Exec(ctx, q,
			p.ID, p.Title, p.Handle, p.BodyHTML, p.Vendor, p.ProductType, p.Tags,
			string(p.Status), p.Options, p.Variants, p.Images, p.CreatedAt, p.UpdatedAt, p.PublishedAt,
		); err != nil {
			return fmt.Errorf("seed product %q: %w", p.Handle, err)
		}
	}

	// The seed uses explicit IDs, so advance the identity sequence to avoid
	// collisions with future inserts.
	if _, err := r.pool.Exec(ctx,
		`SELECT setval(pg_get_serial_sequence('products', 'id'), (SELECT MAX(id) FROM products))`,
	); err != nil {
		return fmt.Errorf("advance product id sequence: %w", err)
	}
	return nil
}
