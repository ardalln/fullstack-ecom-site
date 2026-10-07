package postgres

import (
	"context"
	"fmt"

	"shop-api/internal/domain"
)

type DashboardRepository struct{ db DBTX }

func NewDashboardRepository(db DBTX) *DashboardRepository { return &DashboardRepository{db: db} }

func (r *DashboardRepository) GetDashboardStats(ctx context.Context) (domain.DashboardStats, error) {
	var stats domain.DashboardStats
	err := r.db.QueryRow(ctx, `SELECT
		(SELECT COUNT(*) FROM products),
		(SELECT COUNT(*) FROM categories),
		(SELECT COUNT(*) FROM users),
		(SELECT COUNT(*) FROM orders),
		(SELECT COUNT(*) FROM orders WHERE status = 'awaiting_payment'),
		(SELECT COALESCE(SUM(total_amount), 0) FROM orders WHERE status IN ('paid', 'processing', 'shipped'))`).
		Scan(&stats.Products, &stats.Categories, &stats.Users, &stats.Orders, &stats.PendingOrders, &stats.PaidRevenue)
	if err != nil {
		return stats, fmt.Errorf("read admin dashboard stats: %w", err)
	}
	return stats, nil
}
