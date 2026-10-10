package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"shop-api/internal/domain"
)

type ProductRepository struct{ db DBTX }

func NewProductRepository(db DBTX) *ProductRepository { return &ProductRepository{db: db} }

const productColumns = `p.id, p.name, p.slug, p.description, p.image_url, p.image_urls, p.attributes,
	p.variants, p.variant_label, p.category_id, c.name, c.slug, p.brand_id, b.name, b.slug, p.price, p.discount_price, p.stock, p.is_active, p.is_popular, p.created_at, p.updated_at`
const productFrom = ` FROM products p LEFT JOIN categories c ON c.id = p.category_id JOIN brands b ON b.id = p.brand_id `

func scanProduct(row pgx.Row) (*domain.Product, error) {
	var p domain.Product
	var categoryID pgtype.Int8
	var categoryName pgtype.Text
	var categorySlug pgtype.Text
	var imageURLs []string
	var attributes []byte
	var variants []byte
	if err := row.Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.ImageURL, &imageURLs, &attributes,
		&variants, &p.VariantLabel, &categoryID, &categoryName, &categorySlug, &p.BrandID, &p.BrandName, &p.BrandSlug, &p.Price, &p.DiscountPrice, &p.Stock, &p.IsActive, &p.IsPopular, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, err
	}
	p.ImageURLs = imageURLs
	p.Attributes = map[string]string{}
	if len(attributes) > 0 {
		if err := json.Unmarshal(attributes, &p.Attributes); err != nil {
			return nil, fmt.Errorf("decode product attributes: %w", err)
		}
	}
	p.Variants = []domain.ProductVariant{}
	if len(variants) > 0 {
		if err := json.Unmarshal(variants, &p.Variants); err != nil {
			return nil, fmt.Errorf("decode product variants: %w", err)
		}
	}
	if len(p.Variants) > 0 {
		p.Price = p.Variants[0].Price
		p.DiscountPrice = 0
		p.Stock = 0
		for _, variant := range p.Variants {
			if variant.Price < p.Price {
				p.Price = variant.Price
			}
			p.Stock += variant.Stock
		}
	}
	if p.ImageURL == "" && len(p.ImageURLs) > 0 {
		p.ImageURL = p.ImageURLs[0]
	}
	if categoryID.Valid {
		id := categoryID.Int64
		p.CategoryID = &id
	}
	if categoryName.Valid {
		name := categoryName.String
		p.CategoryName = &name
	}
	if categorySlug.Valid {
		p.CategorySlug = categorySlug.String
	}
	return &p, nil
}

func productVariants(p *domain.Product) (string, error) {
	if p.Variants == nil {
		p.Variants = []domain.ProductVariant{}
	}
	data, err := json.Marshal(p.Variants)
	if err != nil {
		return "", fmt.Errorf("encode product variants: %w", err)
	}
	return string(data), nil
}

func productAttributes(p *domain.Product) (string, error) {
	if p.Attributes == nil {
		p.Attributes = map[string]string{}
	}
	data, err := json.Marshal(p.Attributes)
	if err != nil {
		return "", fmt.Errorf("encode product attributes: %w", err)
	}
	return string(data), nil
}

func (r *ProductRepository) Create(ctx context.Context, p *domain.Product) error {
	attrs, err := productAttributes(p)
	if err != nil {
		return err
	}
	variants, err := productVariants(p)
	if err != nil {
		return err
	}
	const q = `INSERT INTO products (name, slug, description, image_url, image_urls, attributes, variants, variant_label, category_id, brand_id, price, discount_price, stock, is_active, is_popular)
		VALUES ($1,$2,$3,$4,$5,$6::jsonb,$7::jsonb,$8,$9,$10,$11,$12,$13,$14,$15) RETURNING id, created_at, updated_at`
	err = r.db.QueryRow(ctx, q, p.Name, p.Slug, p.Description, p.ImageURL, p.ImageURLs, attrs, variants, p.VariantLabel, p.CategoryID, p.BrandID, p.Price, p.DiscountPrice, p.Stock, p.IsActive, p.IsPopular).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if hasPgCode(err, pgForeignKeyViolation) {
			return domain.ErrNotFound
		}
		if hasPgCode(err, pgUniqueViolation) {
			constraint, _ := pgConstraint(err)
			if constraint == "idx_products_slug" {
				return domain.ErrProductSlugTaken
			}
		}
		return fmt.Errorf("insert product: %w", err)
	}
	return nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id int64) (*domain.Product, error) {
	return r.get(ctx, `WHERE p.id = $1`, id)
}

func (r *ProductRepository) GetBySlug(ctx context.Context, slug string) (*domain.Product, error) {
	return r.get(ctx, `WHERE p.slug = $1`, slug)
}

func (r *ProductRepository) get(ctx context.Context, where string, arg any) (*domain.Product, error) {
	p, err := scanProduct(r.db.QueryRow(ctx, `SELECT `+productColumns+productFrom+where, arg))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get product: %w", err)
	}
	return p, nil
}

func (r *ProductRepository) List(ctx context.Context, search string, categoryID, brandID *int64, limit, offset int) ([]domain.Product, int, error) {
	return r.list(ctx, search, categoryID, brandID, false, false, limit, offset)
}

func (r *ProductRepository) ListPopular(ctx context.Context, limit, offset int) ([]domain.Product, int, error) {
	return r.list(ctx, "", nil, nil, false, true, limit, offset)
}

func (r *ProductRepository) ListAdmin(ctx context.Context, search string, categoryID, brandID *int64, limit, offset int) ([]domain.Product, int, error) {
	return r.list(ctx, search, categoryID, brandID, true, false, limit, offset)
}

func (r *ProductRepository) list(ctx context.Context, search string, categoryID, brandID *int64, includeInactive, popularOnly bool, limit, offset int) ([]domain.Product, int, error) {
	pattern := "%" + search + "%"
	const where = `($1::bigint IS NULL OR p.category_id IN (SELECT id FROM category_tree))
		AND ($2::bigint IS NULL OR p.brand_id = $2)
		AND ($3 = '' OR p.name ILIKE $4 OR p.slug ILIKE $4 OR p.description ILIKE $4 OR b.name ILIKE $4)
		AND ($5::boolean OR p.is_active)
		AND (NOT $6::boolean OR p.is_popular)`
	const categoryTree = `WITH RECURSIVE category_tree(id) AS (
		SELECT id FROM categories WHERE $1::bigint IS NOT NULL AND id = $1
		UNION ALL SELECT c.id FROM categories c JOIN category_tree tree ON c.parent_id = tree.id
	) `
	var total int
	if err := r.db.QueryRow(ctx, categoryTree+`SELECT COUNT(*) FROM products p JOIN brands b ON b.id=p.brand_id WHERE `+where, categoryID, brandID, search, pattern, includeInactive, popularOnly).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count products: %w", err)
	}
	rows, err := r.db.Query(ctx, categoryTree+`SELECT `+productColumns+productFrom+`WHERE `+where+` ORDER BY p.id DESC LIMIT $7 OFFSET $8`, categoryID, brandID, search, pattern, includeInactive, popularOnly, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list products: %w", err)
	}
	defer rows.Close()
	products := make([]domain.Product, 0, limit)
	for rows.Next() {
		product, err := scanProduct(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan product: %w", err)
		}
		products = append(products, *product)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate products: %w", err)
	}
	return products, total, nil
}

func (r *ProductRepository) Update(ctx context.Context, p *domain.Product) error {
	attrs, err := productAttributes(p)
	if err != nil {
		return err
	}
	variants, err := productVariants(p)
	if err != nil {
		return err
	}
	variantIDs := make([]string, 0, len(p.Variants))
	for _, variant := range p.Variants {
		variantIDs = append(variantIDs, variant.ID)
	}
	var hasPendingVariantOrder bool
	if err := r.db.QueryRow(ctx, `SELECT EXISTS (
		SELECT 1 FROM order_items oi JOIN orders o ON o.id = oi.order_id
		WHERE oi.product_id = $1 AND oi.variant_id <> '' AND o.status = 'awaiting_payment'
		AND NOT (oi.variant_id = ANY($2::text[])))`, p.ID, variantIDs).Scan(&hasPendingVariantOrder); err != nil {
		return fmt.Errorf("check pending orders before changing variants: %w", err)
	}
	if hasPendingVariantOrder {
		return domain.ErrProductVariantInUse
	}
	const q = `UPDATE products SET name=$2, slug=$3, description=$4, image_url=$5, image_urls=$6,
		attributes=$7::jsonb, variants=$8::jsonb, variant_label=$9, category_id=$10, brand_id=$11, price=$12, discount_price=$13, stock=$14, is_active=$15, is_popular=$16, updated_at=now()
		WHERE id=$1 RETURNING created_at, updated_at`
	err = r.db.QueryRow(ctx, q, p.ID, p.Name, p.Slug, p.Description, p.ImageURL, p.ImageURLs, attrs, variants, p.VariantLabel, p.CategoryID, p.BrandID, p.Price, p.DiscountPrice, p.Stock, p.IsActive, p.IsPopular).
		Scan(&p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	if hasPgCode(err, pgForeignKeyViolation) {
		return domain.ErrNotFound
	}
	if hasPgCode(err, pgUniqueViolation) {
		constraint, _ := pgConstraint(err)
		if constraint == "idx_products_slug" {
			return domain.ErrProductSlugTaken
		}
	}
	if err != nil {
		return fmt.Errorf("update product: %w", err)
	}
	return nil
}

func (r *ProductRepository) SetActive(ctx context.Context, id int64, active bool) error {
	tag, err := r.db.Exec(ctx, `UPDATE products SET is_active=$2, updated_at=now() WHERE id=$1`, id, active)
	if err != nil {
		return fmt.Errorf("set product visibility: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ProductRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		if hasPgCode(err, pgForeignKeyViolation) {
			return domain.ErrProductInUse
		}
		return fmt.Errorf("delete product: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// GetByIDsForUpdate locks rows in id order to keep concurrent checkouts deadlock-safe.
func (r *ProductRepository) GetByIDsForUpdate(ctx context.Context, ids []int64) ([]domain.Product, error) {
	rows, err := r.db.Query(ctx, `SELECT `+productColumns+productFrom+`WHERE p.id = ANY($1) ORDER BY p.id FOR UPDATE OF p`, ids)
	if err != nil {
		return nil, fmt.Errorf("lock products: %w", err)
	}
	defer rows.Close()
	products := make([]domain.Product, 0, len(ids))
	for rows.Next() {
		product, err := scanProduct(rows)
		if err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}
		products = append(products, *product)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate locked products: %w", err)
	}
	return products, nil
}

func (r *ProductRepository) DecreaseStock(ctx context.Context, id int64, qty int) error {
	tag, err := r.db.Exec(ctx, `UPDATE products SET stock = stock - $2, updated_at = now() WHERE id = $1 AND stock >= $2`, id, qty)
	if err != nil {
		return fmt.Errorf("decrease stock: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrInsufficientStock
	}
	return nil
}

func (r *ProductRepository) IncreaseStock(ctx context.Context, id int64, qty int) error {
	tag, err := r.db.Exec(ctx, `UPDATE products SET stock = stock + $2, updated_at = now() WHERE id = $1`, id, qty)
	if err != nil {
		return fmt.Errorf("increase stock: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ProductRepository) DecreaseVariantStock(ctx context.Context, productID int64, variantID string, qty int) error {
	const q = `UPDATE products p SET
		variants = (SELECT jsonb_agg(CASE WHEN v.item->>'id' = $2
			THEN jsonb_set(v.item, '{stock}', to_jsonb((v.item->>'stock')::integer - $3), false)
			ELSE v.item END ORDER BY v.ordinality)
			FROM jsonb_array_elements(p.variants) WITH ORDINALITY AS v(item, ordinality)),
		stock = p.stock - $3, updated_at = now()
		WHERE p.id = $1 AND EXISTS (SELECT 1 FROM jsonb_array_elements(p.variants) AS v(item)
			WHERE v.item->>'id' = $2 AND (v.item->>'stock')::integer >= $3)`
	tag, err := r.db.Exec(ctx, q, productID, variantID, qty)
	if err != nil {
		return fmt.Errorf("decrease product variant stock: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrInsufficientStock
	}
	return nil
}

func (r *ProductRepository) IncreaseVariantStock(ctx context.Context, productID int64, variantID string, qty int) error {
	const q = `UPDATE products p SET
		variants = (SELECT jsonb_agg(CASE WHEN v.item->>'id' = $2
			THEN jsonb_set(v.item, '{stock}', to_jsonb((v.item->>'stock')::integer + $3), false)
			ELSE v.item END ORDER BY v.ordinality)
			FROM jsonb_array_elements(p.variants) WITH ORDINALITY AS v(item, ordinality)),
		stock = p.stock + $3, updated_at = now()
		WHERE p.id = $1 AND EXISTS (SELECT 1 FROM jsonb_array_elements(p.variants) AS v(item)
			WHERE v.item->>'id' = $2)`
	tag, err := r.db.Exec(ctx, q, productID, variantID, qty)
	if err != nil {
		return fmt.Errorf("increase product variant stock: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
