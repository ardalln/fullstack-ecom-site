package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"shop-api/internal/domain"
)

type BlogRepository struct{ db DBTX }

func NewBlogRepository(db DBTX) *BlogRepository { return &BlogRepository{db: db} }

const blogPostColumns = `id, title, slug, summary, content, cover_image_url, seo_title, seo_description, is_published, published_at, created_at, updated_at`
const blogPostListColumns = `id, title, slug, summary, ''::text, cover_image_url, seo_title, seo_description, is_published, published_at, created_at, updated_at`

func scanBlogPost(row pgx.Row) (*domain.BlogPost, error) {
	var post domain.BlogPost
	var publishedAt pgtype.Timestamptz
	if err := row.Scan(&post.ID, &post.Title, &post.Slug, &post.Summary, &post.Content, &post.CoverImageURL,
		&post.SEOTitle, &post.SEODescription, &post.IsPublished, &publishedAt, &post.CreatedAt, &post.UpdatedAt); err != nil {
		return nil, err
	}
	if publishedAt.Valid {
		value := publishedAt.Time
		post.PublishedAt = &value
	}
	return &post, nil
}

func (r *BlogRepository) Create(ctx context.Context, post *domain.BlogPost) error {
	var publishedAt pgtype.Timestamptz
	err := r.db.QueryRow(ctx, `INSERT INTO blog_posts (title, slug, summary, content, cover_image_url, seo_title, seo_description, is_published, published_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,CASE WHEN $8 THEN now() ELSE NULL END)
		RETURNING id, published_at, created_at, updated_at`, post.Title, post.Slug, post.Summary, post.Content,
		post.CoverImageURL, post.SEOTitle, post.SEODescription, post.IsPublished).
		Scan(&post.ID, &publishedAt, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		if hasPgCode(err, pgUniqueViolation) {
			return domain.ErrBlogSlugTaken
		}
		return fmt.Errorf("insert blog post: %w", err)
	}
	if publishedAt.Valid {
		value := publishedAt.Time
		post.PublishedAt = &value
	}
	return nil
}

func (r *BlogRepository) GetByID(ctx context.Context, id int64) (*domain.BlogPost, error) {
	return r.get(ctx, `WHERE id=$1`, id)
}

func (r *BlogRepository) GetBySlug(ctx context.Context, slug string) (*domain.BlogPost, error) {
	return r.get(ctx, `WHERE slug=$1`, slug)
}

func (r *BlogRepository) get(ctx context.Context, where string, arg any) (*domain.BlogPost, error) {
	post, err := scanBlogPost(r.db.QueryRow(ctx, `SELECT `+blogPostColumns+` FROM blog_posts `+where, arg))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get blog post: %w", err)
	}
	return post, nil
}

func (r *BlogRepository) ListPublished(ctx context.Context, search string, limit, offset int) ([]domain.BlogPost, int, error) {
	return r.list(ctx, search, false, limit, offset)
}

func (r *BlogRepository) ListAdmin(ctx context.Context, search string, limit, offset int) ([]domain.BlogPost, int, error) {
	return r.list(ctx, search, true, limit, offset)
}

func (r *BlogRepository) list(ctx context.Context, search string, includeDrafts bool, limit, offset int) ([]domain.BlogPost, int, error) {
	pattern := "%" + search + "%"
	const where = `($1::boolean OR is_published) AND ($2='' OR title ILIKE $3 OR slug ILIKE $3 OR summary ILIKE $3)`
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM blog_posts WHERE `+where, includeDrafts, search, pattern).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count blog posts: %w", err)
	}
	order := `published_at DESC NULLS LAST, id DESC`
	if includeDrafts {
		order = `updated_at DESC, id DESC`
	}
	rows, err := r.db.Query(ctx, `SELECT `+blogPostListColumns+` FROM blog_posts WHERE `+where+` ORDER BY `+order+` LIMIT $4 OFFSET $5`, includeDrafts, search, pattern, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list blog posts: %w", err)
	}
	defer rows.Close()
	posts := make([]domain.BlogPost, 0, limit)
	for rows.Next() {
		post, err := scanBlogPost(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan blog post: %w", err)
		}
		posts = append(posts, *post)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate blog posts: %w", err)
	}
	return posts, total, nil
}

func (r *BlogRepository) Update(ctx context.Context, post *domain.BlogPost) error {
	var publishedAt pgtype.Timestamptz
	err := r.db.QueryRow(ctx, `UPDATE blog_posts SET title=$2, slug=$3, summary=$4, content=$5, cover_image_url=$6,
		seo_title=$7, seo_description=$8, is_published=$9,
		published_at=CASE WHEN $9 THEN COALESCE(published_at, now()) ELSE NULL END, updated_at=now()
		WHERE id=$1 RETURNING published_at, created_at, updated_at`, post.ID, post.Title, post.Slug, post.Summary,
		post.Content, post.CoverImageURL, post.SEOTitle, post.SEODescription, post.IsPublished).
		Scan(&publishedAt, &post.CreatedAt, &post.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	if hasPgCode(err, pgUniqueViolation) {
		return domain.ErrBlogSlugTaken
	}
	if err != nil {
		return fmt.Errorf("update blog post: %w", err)
	}
	if publishedAt.Valid {
		value := publishedAt.Time
		post.PublishedAt = &value
	}
	return nil
}

func (r *BlogRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM blog_posts WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("delete blog post: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
