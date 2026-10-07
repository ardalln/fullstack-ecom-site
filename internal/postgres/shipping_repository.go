package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"shop-api/internal/domain"
)

type ShippingRepository struct{ db DBTX }

func NewShippingRepository(db DBTX) *ShippingRepository { return &ShippingRepository{db: db} }

const shippingColumns = `id,name,description,province,city,price,is_active,sort_order,created_at,updated_at`

func scanShipping(row pgx.Row) (*domain.ShippingMethod, error) {
	var m domain.ShippingMethod
	err := row.Scan(&m.ID, &m.Name, &m.Description, &m.Province, &m.City, &m.Price, &m.IsActive, &m.SortOrder, &m.CreatedAt, &m.UpdatedAt)
	return &m, err
}
func (r *ShippingRepository) ListAvailable(ctx context.Context, province, city string) ([]domain.ShippingMethod, error) {
	rows, err := r.db.Query(ctx, `SELECT `+shippingColumns+` FROM shipping_methods WHERE is_active AND (province='' OR province=$1) AND (city='' OR city=$2) ORDER BY (province<>'') DESC,(city<>'') DESC,sort_order,id`, province, city)
	if err != nil {
		return nil, fmt.Errorf("list shipping methods: %w", err)
	}
	defer rows.Close()
	out := []domain.ShippingMethod{}
	for rows.Next() {
		m, e := scanShipping(rows)
		if e != nil {
			return nil, fmt.Errorf("scan shipping method: %w", e)
		}
		out = append(out, *m)
	}
	if e := rows.Err(); e != nil {
		return nil, fmt.Errorf("iterate shipping methods: %w", e)
	}
	return out, nil
}
func (r *ShippingRepository) ListAll(ctx context.Context) ([]domain.ShippingMethod, error) {
	rows, err := r.db.Query(ctx, `SELECT `+shippingColumns+` FROM shipping_methods ORDER BY sort_order,id`)
	if err != nil {
		return nil, fmt.Errorf("list shipping methods: %w", err)
	}
	defer rows.Close()
	out := []domain.ShippingMethod{}
	for rows.Next() {
		m, e := scanShipping(rows)
		if e != nil {
			return nil, fmt.Errorf("scan shipping method: %w", e)
		}
		out = append(out, *m)
	}
	if e := rows.Err(); e != nil {
		return nil, fmt.Errorf("iterate shipping methods: %w", e)
	}
	return out, nil
}
func (r *ShippingRepository) GetByID(ctx context.Context, id int64) (*domain.ShippingMethod, error) {
	m, err := scanShipping(r.db.QueryRow(ctx, `SELECT `+shippingColumns+` FROM shipping_methods WHERE id=$1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get shipping method: %w", err)
	}
	return m, nil
}
func (r *ShippingRepository) Create(ctx context.Context, m *domain.ShippingMethod) error {
	return r.db.QueryRow(ctx, `INSERT INTO shipping_methods(name,description,province,city,price,is_active,sort_order) VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING id,created_at,updated_at`, m.Name, m.Description, m.Province, m.City, m.Price, m.IsActive, m.SortOrder).Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt)
}
func (r *ShippingRepository) Update(ctx context.Context, m *domain.ShippingMethod) error {
	err := r.db.QueryRow(ctx, `UPDATE shipping_methods SET name=$2,description=$3,province=$4,city=$5,price=$6,is_active=$7,sort_order=$8,updated_at=now() WHERE id=$1 RETURNING created_at,updated_at`, m.ID, m.Name, m.Description, m.Province, m.City, m.Price, m.IsActive, m.SortOrder).Scan(&m.CreatedAt, &m.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("update shipping method: %w", err)
	}
	return nil
}
func (r *ShippingRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM shipping_methods WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("delete shipping method: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
