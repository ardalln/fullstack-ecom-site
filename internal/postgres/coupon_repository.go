package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"shop-api/internal/domain"
)

type CouponRepository struct{ db DBTX }

func NewCouponRepository(db DBTX) *CouponRepository { return &CouponRepository{db: db} }

const couponColumns = `id,code,kind,value,minimum_subtotal,maximum_discount,usage_limit,used_count,starts_at,ends_at,is_active,created_at,updated_at`

func scanCoupon(row pgx.Row) (*domain.Coupon, error) {
	var c domain.Coupon
	err := row.Scan(&c.ID, &c.Code, &c.Kind, &c.Value, &c.MinimumSubtotal, &c.MaximumDiscount, &c.UsageLimit, &c.UsedCount, &c.StartsAt, &c.EndsAt, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	return &c, err
}
func (r *CouponRepository) GetByCode(ctx context.Context, code string) (*domain.Coupon, error) {
	return r.get(ctx, `SELECT `+couponColumns+` FROM coupons WHERE lower(code)=lower($1)`, code)
}
func (r *CouponRepository) GetByCodeForUpdate(ctx context.Context, code string) (*domain.Coupon, error) {
	return r.get(ctx, `SELECT `+couponColumns+` FROM coupons WHERE lower(code)=lower($1) FOR UPDATE`, code)
}
func (r *CouponRepository) get(ctx context.Context, q string, arg any) (*domain.Coupon, error) {
	c, err := scanCoupon(r.db.QueryRow(ctx, q, arg))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get coupon: %w", err)
	}
	return c, nil
}
func (r *CouponRepository) ListAll(ctx context.Context) ([]domain.Coupon, error) {
	rows, err := r.db.Query(ctx, `SELECT `+couponColumns+` FROM coupons ORDER BY id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list coupons: %w", err)
	}
	defer rows.Close()
	out := []domain.Coupon{}
	for rows.Next() {
		c, e := scanCoupon(rows)
		if e != nil {
			return nil, fmt.Errorf("scan coupon: %w", e)
		}
		out = append(out, *c)
	}
	if e := rows.Err(); e != nil {
		return nil, fmt.Errorf("iterate coupons: %w", e)
	}
	return out, nil
}
func (r *CouponRepository) Create(ctx context.Context, c *domain.Coupon) error {
	err := r.db.QueryRow(ctx, `INSERT INTO coupons(code,kind,value,minimum_subtotal,maximum_discount,usage_limit,starts_at,ends_at,is_active) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id,used_count,created_at,updated_at`, c.Code, c.Kind, c.Value, c.MinimumSubtotal, c.MaximumDiscount, c.UsageLimit, c.StartsAt, c.EndsAt, c.IsActive).Scan(&c.ID, &c.UsedCount, &c.CreatedAt, &c.UpdatedAt)
	if hasPgCode(err, pgUniqueViolation) {
		return fmt.Errorf("%w: coupon code is already in use", domain.ErrInvalidInput)
	}
	if err != nil {
		return fmt.Errorf("create coupon: %w", err)
	}
	return nil
}
func (r *CouponRepository) Update(ctx context.Context, c *domain.Coupon) error {
	err := r.db.QueryRow(ctx, `UPDATE coupons SET code=$2,kind=$3,value=$4,minimum_subtotal=$5,maximum_discount=$6,usage_limit=$7,starts_at=$8,ends_at=$9,is_active=$10,updated_at=now() WHERE id=$1 RETURNING used_count,created_at,updated_at`, c.ID, c.Code, c.Kind, c.Value, c.MinimumSubtotal, c.MaximumDiscount, c.UsageLimit, c.StartsAt, c.EndsAt, c.IsActive).Scan(&c.UsedCount, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	if hasPgCode(err, pgUniqueViolation) {
		return fmt.Errorf("%w: coupon code is already in use", domain.ErrInvalidInput)
	}
	if err != nil {
		return fmt.Errorf("update coupon: %w", err)
	}
	return nil
}
func (r *CouponRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM coupons WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("delete coupon: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
func (r *CouponRepository) ReserveUse(ctx context.Context, id int64, now time.Time) error {
	tag, err := r.db.Exec(ctx, `UPDATE coupons SET used_count=used_count+1,updated_at=now() WHERE id=$1 AND is_active AND (starts_at IS NULL OR starts_at<=$2) AND (ends_at IS NULL OR ends_at>$2) AND (usage_limit IS NULL OR used_count<usage_limit)`, id, now)
	if err != nil {
		return fmt.Errorf("reserve coupon use: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrCouponInvalid
	}
	return nil
}
func (r *CouponRepository) ReleaseUse(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `UPDATE coupons SET used_count=GREATEST(0,used_count-1),updated_at=now() WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("release coupon use: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
