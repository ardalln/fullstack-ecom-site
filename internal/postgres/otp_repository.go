package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"shop-api/internal/domain"
)

type OTPRepository struct {
	db DBTX
}

func NewOTPRepository(db DBTX) *OTPRepository {
	return &OTPRepository{db: db}
}

func (r *OTPRepository) Upsert(ctx context.Context, phone, codeHash string, expiresAt time.Time) error {
	const q = `
		INSERT INTO otp_codes (phone, code_hash, attempts, expires_at, last_sent_at)
		VALUES ($1, $2, 0, $3, now())
		ON CONFLICT (phone) DO UPDATE
		SET code_hash = EXCLUDED.code_hash,
		    attempts = 0,
		    expires_at = EXCLUDED.expires_at,
		    last_sent_at = now()`
	if _, err := r.db.Exec(ctx, q, phone, codeHash, expiresAt); err != nil {
		return fmt.Errorf("upsert otp code: %w", err)
	}
	return nil
}

func (r *OTPRepository) Get(ctx context.Context, phone string) (*domain.OTPCode, error) {
	const q = `SELECT phone, code_hash, attempts, expires_at, last_sent_at FROM otp_codes WHERE phone = $1`
	var c domain.OTPCode
	err := r.db.QueryRow(ctx, q, phone).Scan(&c.Phone, &c.CodeHash, &c.Attempts, &c.ExpiresAt, &c.LastSentAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get otp code: %w", err)
	}
	return &c, nil
}

func (r *OTPRepository) IncrementAttempts(ctx context.Context, phone string) error {
	if _, err := r.db.Exec(ctx, `UPDATE otp_codes SET attempts = attempts + 1 WHERE phone = $1`, phone); err != nil {
		return fmt.Errorf("increment otp attempts: %w", err)
	}
	return nil
}

func (r *OTPRepository) Delete(ctx context.Context, phone string) error {
	if _, err := r.db.Exec(ctx, `DELETE FROM otp_codes WHERE phone = $1`, phone); err != nil {
		return fmt.Errorf("delete otp code: %w", err)
	}
	return nil
}
