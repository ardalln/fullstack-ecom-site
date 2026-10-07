package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"shop-api/internal/domain"
)

type BrandRepository struct{ db DBTX }

func NewBrandRepository(db DBTX) *BrandRepository { return &BrandRepository{db: db} }

const brandColumns = `id, name, slug, description, seo_title, seo_description, created_at, updated_at`

func scanBrand(row pgx.Row) (*domain.Brand, error) {
	var brand domain.Brand
	if err := row.Scan(&brand.ID, &brand.Name, &brand.Slug, &brand.Description, &brand.SEOTitle, &brand.SEODescription, &brand.CreatedAt, &brand.UpdatedAt); err != nil {
		return nil, err
	}
	return &brand, nil
}

func (r *BrandRepository) Create(ctx context.Context, brand *domain.Brand) error {
	err := r.db.QueryRow(ctx, `INSERT INTO brands (name, slug, description, seo_title, seo_description) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at`, brand.Name, brand.Slug, brand.Description, brand.SEOTitle, brand.SEODescription).
		Scan(&brand.ID, &brand.CreatedAt, &brand.UpdatedAt)
	if err != nil {
		if hasPgCode(err, pgUniqueViolation) {
			constraint, _ := pgConstraint(err)
			if constraint == "idx_brands_slug" {
				return domain.ErrBrandSlugTaken
			}
			return domain.ErrBrandNameTaken
		}
		return fmt.Errorf("insert brand: %w", err)
	}
	return nil
}

func (r *BrandRepository) GetBySlug(ctx context.Context, slug string) (*domain.Brand, error) {
	brand, err := scanBrand(r.db.QueryRow(ctx, `SELECT `+brandColumns+` FROM brands WHERE slug = $1`, slug))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get brand by slug: %w", err)
	}
	return brand, nil
}

func (r *BrandRepository) GetByID(ctx context.Context, id int64) (*domain.Brand, error) {
	brand, err := scanBrand(r.db.QueryRow(ctx, `SELECT `+brandColumns+` FROM brands WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get brand: %w", err)
	}
	return brand, nil
}

func (r *BrandRepository) List(ctx context.Context) ([]domain.Brand, error) {
	rows, err := r.db.Query(ctx, `SELECT `+brandColumns+` FROM brands ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list brands: %w", err)
	}
	defer rows.Close()
	brands := make([]domain.Brand, 0)
	for rows.Next() {
		brand, err := scanBrand(rows)
		if err != nil {
			return nil, fmt.Errorf("scan brand: %w", err)
		}
		brands = append(brands, *brand)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate brands: %w", err)
	}
	return brands, nil
}

func (r *BrandRepository) Update(ctx context.Context, brand *domain.Brand) error {
	err := r.db.QueryRow(ctx, `UPDATE brands SET name = $2, slug = $3, description = $4, seo_title = $5, seo_description = $6, updated_at = now() WHERE id = $1 RETURNING created_at, updated_at`, brand.ID, brand.Name, brand.Slug, brand.Description, brand.SEOTitle, brand.SEODescription).
		Scan(&brand.CreatedAt, &brand.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	if hasPgCode(err, pgUniqueViolation) {
		constraint, _ := pgConstraint(err)
		if constraint == "idx_brands_slug" {
			return domain.ErrBrandSlugTaken
		}
		return domain.ErrBrandNameTaken
	}
	if err != nil {
		return fmt.Errorf("update brand: %w", err)
	}
	return nil
}

func (r *BrandRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM brands WHERE id = $1`, id)
	if err != nil {
		if hasPgCode(err, pgForeignKeyViolation) {
			return domain.ErrBrandInUse
		}
		return fmt.Errorf("delete brand: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
