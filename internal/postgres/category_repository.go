package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"shop-api/internal/domain"
)

type CategoryRepository struct{ db DBTX }

func NewCategoryRepository(db DBTX) *CategoryRepository { return &CategoryRepository{db: db} }

const categoryColumns = `c.id, c.name, c.slug, c.description, c.seo_title, c.seo_description, c.home_title, c.home_image_url, c.show_on_home, c.parent_id, p.name, c.created_at, c.updated_at`

func scanCategory(row pgx.Row) (*domain.Category, error) {
	var category domain.Category
	var parentID pgtype.Int8
	var parentName pgtype.Text
	if err := row.Scan(&category.ID, &category.Name, &category.Slug, &category.Description, &category.SEOTitle, &category.SEODescription, &category.HomeTitle, &category.HomeImageURL, &category.ShowOnHome, &parentID, &parentName, &category.CreatedAt, &category.UpdatedAt); err != nil {
		return nil, err
	}
	if parentID.Valid {
		id := parentID.Int64
		category.ParentID = &id
	}
	if parentName.Valid {
		name := parentName.String
		category.ParentName = &name
	}
	return &category, nil
}

func (r *CategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	err := r.db.QueryRow(ctx, `INSERT INTO categories (name, slug, description, seo_title, seo_description, home_title, home_image_url, show_on_home, parent_id) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id, created_at, updated_at`, category.Name, category.Slug, category.Description, category.SEOTitle, category.SEODescription, category.HomeTitle, category.HomeImageURL, category.ShowOnHome, category.ParentID).
		Scan(&category.ID, &category.CreatedAt, &category.UpdatedAt)
	if err != nil {
		if hasPgCode(err, pgUniqueViolation) {
			constraint, _ := pgConstraint(err)
			if constraint == "idx_categories_slug" {
				return domain.ErrCategorySlugTaken
			}
			return domain.ErrCategoryNameTaken
		}
		return fmt.Errorf("insert category: %w", err)
	}
	return nil
}

func (r *CategoryRepository) GetBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	category, err := scanCategory(r.db.QueryRow(ctx, `SELECT `+categoryColumns+` FROM categories c LEFT JOIN categories p ON p.id = c.parent_id WHERE c.slug = $1`, slug))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get category by slug: %w", err)
	}
	return category, nil
}

func (r *CategoryRepository) GetByID(ctx context.Context, id int64) (*domain.Category, error) {
	category, err := scanCategory(r.db.QueryRow(ctx, `SELECT `+categoryColumns+` FROM categories c LEFT JOIN categories p ON p.id = c.parent_id WHERE c.id = $1`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get category: %w", err)
	}
	return category, nil
}

func (r *CategoryRepository) List(ctx context.Context) ([]domain.Category, error) {
	rows, err := r.db.Query(ctx, `SELECT `+categoryColumns+` FROM categories c LEFT JOIN categories p ON p.id = c.parent_id ORDER BY p.name NULLS FIRST, c.name`)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()
	categories := make([]domain.Category, 0)
	for rows.Next() {
		category, err := scanCategory(rows)
		if err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		categories = append(categories, *category)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate categories: %w", err)
	}
	return categories, nil
}

func (r *CategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	err := r.db.QueryRow(ctx, `UPDATE categories SET name=$2, slug=$3, description=$4, seo_title=$5, seo_description=$6, home_title=$7, home_image_url=$8, show_on_home=$9, parent_id=$10, updated_at=now()
		WHERE id = $1 RETURNING created_at, updated_at`, category.ID, category.Name, category.Slug, category.Description, category.SEOTitle, category.SEODescription, category.HomeTitle, category.HomeImageURL, category.ShowOnHome, category.ParentID).
		Scan(&category.CreatedAt, &category.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		if hasPgCode(err, pgUniqueViolation) {
			constraint, _ := pgConstraint(err)
			if constraint == "idx_categories_slug" {
				return domain.ErrCategorySlugTaken
			}
			return domain.ErrCategoryNameTaken
		}
		return fmt.Errorf("update category: %w", err)
	}
	return nil
}

func (r *CategoryRepository) ValidateParent(ctx context.Context, id int64, parentID *int64) error {
	if parentID == nil {
		return nil
	}
	if *parentID <= 0 || *parentID == id {
		return fmt.Errorf("%w: invalid parent category", domain.ErrInvalidInput)
	}
	var valid bool
	err := r.db.QueryRow(ctx, `WITH RECURSIVE descendants(id) AS (
		SELECT id FROM categories WHERE parent_id = $2
		UNION ALL SELECT c.id FROM categories c JOIN descendants d ON c.parent_id = d.id
	) SELECT EXISTS (SELECT 1 FROM categories WHERE id = $1)
	AND ($2::bigint = 0 OR NOT EXISTS (SELECT 1 FROM descendants WHERE id = $1))`, *parentID, id).Scan(&valid)
	if err != nil {
		return fmt.Errorf("validate category parent: %w", err)
	}
	if !valid {
		return fmt.Errorf("%w: invalid parent category", domain.ErrInvalidInput)
	}
	return nil
}

func (r *CategoryRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM categories WHERE id = $1`, id)
	if err != nil {
		if hasPgCode(err, pgForeignKeyViolation) {
			return domain.ErrCategoryInUse
		}
		return fmt.Errorf("delete category: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
