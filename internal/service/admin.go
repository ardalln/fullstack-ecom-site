package service

import (
	"context"
	"errors"
	"fmt"

	"shop-api/internal/domain"
)

type AdminService struct {
	users     domain.UserRepository
	orders    domain.OrderRepository
	dashboard domain.DashboardRepository
	tx        domain.TxManager
}

func NewAdminService(users domain.UserRepository, orders domain.OrderRepository, dashboard domain.DashboardRepository, tx domain.TxManager) *AdminService {
	return &AdminService{users: users, orders: orders, dashboard: dashboard, tx: tx}
}

func (s *AdminService) Overview(ctx context.Context) (domain.DashboardStats, error) {
	return s.dashboard.GetDashboardStats(ctx)
}

func (s *AdminService) ListUsers(ctx context.Context, search string, page, limit int) ([]domain.User, int, error) {
	return s.users.List(ctx, search, limit, (page-1)*limit)
}

func (s *AdminService) ChangeUserAccess(ctx context.Context, actorID, targetID int64, role domain.Role, active bool) (*domain.User, error) {
	if role != domain.RoleUser && role != domain.RoleAdmin {
		return nil, fmt.Errorf("%w: role must be user or admin", domain.ErrInvalidInput)
	}
	if actorID == targetID {
		return nil, fmt.Errorf("%w: administrators cannot change their own access", domain.ErrInvalidInput)
	}
	var updated *domain.User
	err := s.tx.WithinTx(ctx, func(repos domain.Repositories) error {
		if err := repos.Users.UpdateAccess(ctx, targetID, role, active); err != nil {
			return err
		}
		var err error
		updated, err = repos.Users.GetByID(ctx, targetID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *AdminService) ListOrders(ctx context.Context, status domain.OrderStatus, page, limit int) ([]domain.Order, int, error) {
	return s.orders.ListAll(ctx, status, limit, (page-1)*limit)
}

func (s *AdminService) GetOrder(ctx context.Context, orderID int64) (*domain.Order, error) {
	return s.orders.GetByID(ctx, orderID)
}

func (s *AdminService) ChangeOrderStatus(ctx context.Context, orderID int64, next domain.OrderStatus) (*domain.Order, error) {
	if !validAdminOrderStatus(next) {
		return nil, fmt.Errorf("%w: unsupported order status", domain.ErrInvalidInput)
	}
	err := s.tx.WithinTx(ctx, func(repos domain.Repositories) error {
		order, err := repos.Orders.GetByIDForUpdate(ctx, orderID)
		if err != nil {
			return err
		}
		if !validOrderTransition(order.Status, next) {
			return domain.ErrOrderStatusConflict
		}

		if next == domain.OrderStatusCancelled {
			payment, err := repos.Payments.GetByOrderID(ctx, orderID)
			if err == nil {
				if payment.Status != domain.PaymentPending {
					return domain.ErrOrderStatusConflict
				}
				if err := repos.Payments.MarkCancelled(ctx, payment.Authority); err != nil {
					return err
				}
			} else if !errors.Is(err, domain.ErrNotFound) {
				return err
			}
			if order.CouponID != nil {
				if err := repos.Coupons.ReleaseUse(ctx, *order.CouponID); err != nil {
					return err
				}
			}
		}

		if err := repos.Orders.TransitionStatus(ctx, orderID, order.Status, next); err != nil {
			return err
		}
		if next == domain.OrderStatusCancelled {
			for _, item := range order.Items {
				var err error
				if item.VariantID != "" {
					err = repos.Products.IncreaseVariantStock(ctx, item.ProductID, item.VariantID, item.Quantity)
				} else {
					err = repos.Products.IncreaseStock(ctx, item.ProductID, item.Quantity)
				}
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.orders.GetByID(ctx, orderID)
}

func validAdminOrderStatus(status domain.OrderStatus) bool {
	switch status {
	case domain.OrderStatusProcessing, domain.OrderStatusShipped, domain.OrderStatusCancelled:
		return true
	default:
		return false
	}
}

func validOrderTransition(current, next domain.OrderStatus) bool {
	switch current {
	case domain.OrderStatusAwaitingPayment:
		return next == domain.OrderStatusCancelled
	case domain.OrderStatusPaid:
		return next == domain.OrderStatusProcessing
	case domain.OrderStatusProcessing:
		return next == domain.OrderStatusShipped
	case domain.OrderStatusShipped:
		return false
	default:
		return false
	}
}
