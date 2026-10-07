package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"shop-api/internal/domain"
)

type PaymentRepository struct {
	db DBTX
}

func NewPaymentRepository(db DBTX) *PaymentRepository {
	return &PaymentRepository{db: db}
}

const paymentColumns = `id, order_id, authority, amount, status, ref_id, paid_at, created_at`

func scanPayment(row pgx.Row) (*domain.Payment, error) {
	var (
		p      domain.Payment
		status string
	)
	if err := row.Scan(&p.ID, &p.OrderID, &p.Authority, &p.Amount, &status, &p.RefID, &p.PaidAt, &p.CreatedAt); err != nil {
		return nil, err
	}
	p.Status = domain.PaymentStatus(status)
	return &p, nil
}

func (r *PaymentRepository) Create(ctx context.Context, p *domain.Payment) error {
	const q = `
		INSERT INTO payments (order_id, authority, amount, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`

	err := r.db.QueryRow(ctx, q, p.OrderID, p.Authority, p.Amount, string(p.Status)).
		Scan(&p.ID, &p.CreatedAt)
	if err != nil {
		if hasPgCode(err, pgUniqueViolation) {
			return fmt.Errorf("%w: a payment session already exists for this order", domain.ErrOrderNotPayable)
		}
		return fmt.Errorf("insert payment: %w", err)
	}
	return nil
}

func (r *PaymentRepository) GetByAuthority(ctx context.Context, authority string) (*domain.Payment, error) {
	p, err := scanPayment(r.db.QueryRow(ctx, `SELECT `+paymentColumns+` FROM payments WHERE authority = $1`, authority))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get payment by authority: %w", err)
	}
	return p, nil
}

func (r *PaymentRepository) GetByOrderID(ctx context.Context, orderID int64) (*domain.Payment, error) {
	p, err := scanPayment(r.db.QueryRow(ctx, `SELECT `+paymentColumns+` FROM payments WHERE order_id = $1`, orderID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get payment by order: %w", err)
	}
	return p, nil
}

// MarkPaid only succeeds if the payment is still pending, so a second
// confirm call (e.g. a duplicated callback) cannot pay an order twice.
func (r *PaymentRepository) MarkPaid(ctx context.Context, authority, refID string) error {
	const q = `UPDATE payments SET status = 'paid', ref_id = $2, paid_at = now() WHERE authority = $1 AND status = 'pending'`
	tag, err := r.db.Exec(ctx, q, authority, refID)
	if err != nil {
		return fmt.Errorf("mark payment paid: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrPaymentAlreadyProcessed
	}
	return nil
}

func (r *PaymentRepository) MarkCancelled(ctx context.Context, authority string) error {
	const q = `UPDATE payments SET status = 'cancelled' WHERE authority = $1 AND status = 'pending'`
	tag, err := r.db.Exec(ctx, q, authority)
	if err != nil {
		return fmt.Errorf("mark payment cancelled: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrPaymentAlreadyProcessed
	}
	return nil
}
