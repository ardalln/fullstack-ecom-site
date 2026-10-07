package service

import (
	"context"
	"errors"
	"fmt"

	"shop-api/internal/domain"
)

const authorityBytes = 24 // -> 48 hex chars, comparable to a real gateway's authority token

type PaymentService struct {
	orders   domain.OrderRepository
	payments domain.PaymentRepository
	tx       domain.TxManager
}

func NewPaymentService(orders domain.OrderRepository, payments domain.PaymentRepository, tx domain.TxManager) *PaymentService {
	return &PaymentService{orders: orders, payments: payments, tx: tx}
}

// RequestPayment starts (or resumes) a checkout session for an order,
// mirroring the "payment request" step of a real Iranian gateway like
// ZarinPal: the caller gets back an authority token and can redirect the
// shopper to the page that uses it.
func (s *PaymentService) RequestPayment(ctx context.Context, userID, orderID int64) (*domain.Payment, error) {
	var payment *domain.Payment
	err := s.tx.WithinTx(ctx, func(repos domain.Repositories) error {
		order, err := repos.Orders.GetByIDForUpdate(ctx, orderID)
		if err != nil {
			return err
		}
		if order.UserID != userID {
			return domain.ErrNotFound
		}
		if order.Status != domain.OrderStatusAwaitingPayment {
			return domain.ErrOrderNotPayable
		}
		if existing, err := repos.Payments.GetByOrderID(ctx, orderID); err == nil {
			payment = existing // resume an abandoned checkout instead of starting a new one
			return nil
		} else if !errors.Is(err, domain.ErrNotFound) {
			return err
		}

		authority, err := randomToken(authorityBytes)
		if err != nil {
			return fmt.Errorf("generate payment authority: %w", err)
		}
		payment = &domain.Payment{OrderID: orderID, Authority: authority, Amount: order.TotalAmount, Status: domain.PaymentPending}
		return repos.Payments.Create(ctx, payment)
	})
	if err != nil {
		return nil, err
	}
	return payment, nil
}

// GatewayInfo returns what the mock payment page needs to display. It is
// looked up by authority alone, the same way a real gateway's checkout link
// needs no separate merchant login.
func (s *PaymentService) GatewayInfo(ctx context.Context, authority string) (*domain.Payment, error) {
	return s.payments.GetByAuthority(ctx, authority)
}

// ConfirmPayment simulates the bank reporting the outcome. action is "pay"
// or "cancel". On "pay" the order is marked paid; on "cancel" the order is
// marked cancelled and its reserved stock is released back to the shop.
func (s *PaymentService) ConfirmPayment(ctx context.Context, authority, action string) (*domain.Payment, error) {
	payment, err := s.payments.GetByAuthority(ctx, authority)
	if err != nil {
		return nil, err
	}
	if payment.Status != domain.PaymentPending {
		return nil, domain.ErrPaymentAlreadyProcessed
	}
	if action != "pay" && action != "cancel" {
		return nil, fmt.Errorf("%w: action must be \"pay\" or \"cancel\"", domain.ErrInvalidInput)
	}

	err = s.tx.WithinTx(ctx, func(repos domain.Repositories) error {
		order, err := repos.Orders.GetByIDForUpdate(ctx, payment.OrderID)
		if err != nil {
			return err
		}
		if order.Status != domain.OrderStatusAwaitingPayment {
			return domain.ErrPaymentAlreadyProcessed
		}
		if action == "pay" {
			refID, err := randomDigits(12)
			if err != nil {
				return fmt.Errorf("generate payment reference: %w", err)
			}
			if err := repos.Payments.MarkPaid(ctx, authority, refID); err != nil {
				return err
			}
			return repos.Orders.TransitionStatus(ctx, payment.OrderID, domain.OrderStatusAwaitingPayment, domain.OrderStatusPaid)
		}

		if err := repos.Payments.MarkCancelled(ctx, authority); err != nil {
			return err
		}
		if err := repos.Orders.TransitionStatus(ctx, payment.OrderID, domain.OrderStatusAwaitingPayment, domain.OrderStatusCancelled); err != nil {
			return err
		}
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
		if order.CouponID != nil {
			if err := repos.Coupons.ReleaseUse(ctx, *order.CouponID); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.payments.GetByAuthority(ctx, authority)
}
