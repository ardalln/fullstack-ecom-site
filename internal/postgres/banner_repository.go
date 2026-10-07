package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"shop-api/internal/domain"
)

type BannerRepository struct{ db DBTX }

func NewBannerRepository(db DBTX) *BannerRepository { return &BannerRepository{db: db} }

const bannerColumns = `id,title,subtitle,desktop_image_url,mobile_image_url,link_url,button_label,sort_order,is_active,created_at,updated_at`

func scanBanner(row pgx.Row) (*domain.Banner, error) {
	var b domain.Banner
	err := row.Scan(&b.ID, &b.Title, &b.Subtitle, &b.DesktopImageURL, &b.MobileImageURL, &b.LinkURL, &b.ButtonLabel, &b.SortOrder, &b.IsActive, &b.CreatedAt, &b.UpdatedAt)
	return &b, err
}

func (r *BannerRepository) ListActive(ctx context.Context) ([]domain.Banner, error) {
	return r.list(ctx, true)
}
func (r *BannerRepository) ListAll(ctx context.Context) ([]domain.Banner, error) {
	return r.list(ctx, false)
}
func (r *BannerRepository) list(ctx context.Context, activeOnly bool) ([]domain.Banner, error) {
	rows, err := r.db.Query(ctx, `SELECT `+bannerColumns+` FROM banners WHERE NOT $1 OR is_active ORDER BY sort_order,id`, activeOnly)
	if err != nil {
		return nil, fmt.Errorf("list banners: %w", err)
	}
	defer rows.Close()
	items := []domain.Banner{}
	for rows.Next() {
		item, err := scanBanner(rows)
		if err != nil {
			return nil, fmt.Errorf("scan banner: %w", err)
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate banners: %w", err)
	}
	return items, nil
}
func (r *BannerRepository) Create(ctx context.Context, b *domain.Banner) error {
	return r.db.QueryRow(ctx, `INSERT INTO banners(title,subtitle,desktop_image_url,mobile_image_url,link_url,button_label,sort_order,is_active) VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id,created_at,updated_at`, b.Title, b.Subtitle, b.DesktopImageURL, b.MobileImageURL, b.LinkURL, b.ButtonLabel, b.SortOrder, b.IsActive).Scan(&b.ID, &b.CreatedAt, &b.UpdatedAt)
}
func (r *BannerRepository) Update(ctx context.Context, b *domain.Banner) error {
	err := r.db.QueryRow(ctx, `UPDATE banners SET title=$2,subtitle=$3,desktop_image_url=$4,mobile_image_url=$5,link_url=$6,button_label=$7,sort_order=$8,is_active=$9,updated_at=now() WHERE id=$1 RETURNING created_at,updated_at`, b.ID, b.Title, b.Subtitle, b.DesktopImageURL, b.MobileImageURL, b.LinkURL, b.ButtonLabel, b.SortOrder, b.IsActive).Scan(&b.CreatedAt, &b.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("update banner: %w", err)
	}
	return nil
}
func (r *BannerRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM banners WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("delete banner: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
