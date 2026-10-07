package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"shop-api/internal/domain"
)

// TxManager implements domain.TxManager on top of a pgx pool.
type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

// WithinTx begins a transaction, hands fn a set of repositories bound to it,
// and commits if fn succeeds. Any error or panic triggers a rollback.
func (m *TxManager) WithinTx(ctx context.Context, fn func(repos domain.Repositories) error) error {
	tx, err := m.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after a successful commit

	repos := domain.Repositories{
		Users:      NewUserRepository(tx),
		Categories: NewCategoryRepository(tx),
		Brands:     NewBrandRepository(tx),
		Products:   NewProductRepository(tx),
		Orders:     NewOrderRepository(tx),
		Payments:   NewPaymentRepository(tx),
		Tickets:    NewTicketRepository(tx),
		Shipping:   NewShippingRepository(tx),
		Coupons:    NewCouponRepository(tx),
	}
	if err := fn(repos); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
