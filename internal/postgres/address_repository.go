package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"shop-api/internal/domain"
)

type AddressRepository struct {
	db DBTX
}

func NewAddressRepository(db DBTX) *AddressRepository {
	return &AddressRepository{db: db}
}

const addressColumns = `id, user_id, receiver_name, receiver_phone, province, city, address_line, postal_code, is_default, created_at, updated_at`

func scanAddress(row pgx.Row) (*domain.Address, error) {
	var a domain.Address
	err := row.Scan(&a.ID, &a.UserID, &a.ReceiverName, &a.ReceiverPhone, &a.Province, &a.City,
		&a.AddressLine, &a.PostalCode, &a.IsDefault, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AddressRepository) Create(ctx context.Context, a *domain.Address) error {
	const q = `
		INSERT INTO addresses (user_id, receiver_name, receiver_phone, province, city, address_line, postal_code, is_default)
		VALUES ($1, $2, $3, $4, $5, $6, $7, false)
		RETURNING id, is_default, created_at, updated_at`

	err := r.db.QueryRow(ctx, q, a.UserID, a.ReceiverName, a.ReceiverPhone, a.Province, a.City, a.AddressLine, a.PostalCode).
		Scan(&a.ID, &a.IsDefault, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert address: %w", err)
	}
	return nil
}

func (r *AddressRepository) GetByID(ctx context.Context, userID, id int64) (*domain.Address, error) {
	a, err := scanAddress(r.db.QueryRow(ctx,
		`SELECT `+addressColumns+` FROM addresses WHERE id = $1 AND user_id = $2`, id, userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get address: %w", err)
	}
	return a, nil
}

func (r *AddressRepository) ListByUserID(ctx context.Context, userID int64) ([]domain.Address, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+addressColumns+` FROM addresses WHERE user_id = $1 ORDER BY is_default DESC, id DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list addresses: %w", err)
	}
	defer rows.Close()

	addresses := make([]domain.Address, 0)
	for rows.Next() {
		a, err := scanAddress(rows)
		if err != nil {
			return nil, fmt.Errorf("scan address: %w", err)
		}
		addresses = append(addresses, *a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate addresses: %w", err)
	}
	return addresses, nil
}

func (r *AddressRepository) Update(ctx context.Context, a *domain.Address) error {
	const q = `
		UPDATE addresses
		SET receiver_name = $3, receiver_phone = $4, province = $5, city = $6,
		    address_line = $7, postal_code = $8, updated_at = now()
		WHERE id = $1 AND user_id = $2
		RETURNING is_default, created_at, updated_at`

	err := r.db.QueryRow(ctx, q, a.ID, a.UserID, a.ReceiverName, a.ReceiverPhone, a.Province, a.City, a.AddressLine, a.PostalCode).
		Scan(&a.IsDefault, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("update address: %w", err)
	}
	return nil
}

func (r *AddressRepository) Delete(ctx context.Context, userID, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM addresses WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("delete address: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// SetDefault marks one address as default and unmarks the rest in a single
// statement, so the user never ends up with zero or several defaults.
func (r *AddressRepository) SetDefault(ctx context.Context, userID, id int64) error {
	const q = `UPDATE addresses SET is_default = (id = $2), updated_at = now() WHERE user_id = $1`
	tag, err := r.db.Exec(ctx, q, userID, id)
	if err != nil {
		return fmt.Errorf("set default address: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound // user has no addresses at all
	}

	var exists bool
	const check = `SELECT EXISTS (SELECT 1 FROM addresses WHERE id = $1 AND user_id = $2 AND is_default)`
	if err := r.db.QueryRow(ctx, check, id, userID).Scan(&exists); err != nil {
		return fmt.Errorf("verify default address: %w", err)
	}
	if !exists {
		// id did not belong to this user: every address is now non-default.
		return domain.ErrNotFound
	}
	return nil
}
