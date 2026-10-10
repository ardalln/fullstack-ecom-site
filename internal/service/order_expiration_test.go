package service

import (
	"context"
	"testing"
	"time"

	"shop-api/internal/domain"
)

type expirationOrderRepository struct {
	domain.OrderRepository
	order      *domain.Order
	candidates []int64
}

func (r *expirationOrderRepository) ListAwaitingPaymentBefore(context.Context, time.Time, int) ([]int64, error) {
	return r.candidates, nil
}

func (r *expirationOrderRepository) GetByIDForUpdate(context.Context, int64) (*domain.Order, error) {
	return r.order, nil
}

func (r *expirationOrderRepository) TransitionStatus(_ context.Context, _ int64, from, to domain.OrderStatus) error {
	if r.order.Status != from {
		return domain.ErrOrderStatusConflict
	}
	r.order.Status = to
	return nil
}

type expirationPaymentRepository struct {
	domain.PaymentRepository
	payment *domain.Payment
}

func (r *expirationPaymentRepository) GetByOrderID(context.Context, int64) (*domain.Payment, error) {
	if r.payment == nil {
		return nil, domain.ErrNotFound
	}
	return r.payment, nil
}

func (r *expirationPaymentRepository) MarkCancelled(context.Context, string) error {
	r.payment.Status = domain.PaymentCancelled
	return nil
}

type expirationProductRepository struct {
	domain.ProductRepository
	restored []domain.OrderItem
}

func (r *expirationProductRepository) IncreaseStock(_ context.Context, id int64, quantity int) error {
	r.restored = append(r.restored, domain.OrderItem{ProductID: id, Quantity: quantity})
	return nil
}

func (r *expirationProductRepository) IncreaseVariantStock(_ context.Context, id int64, variantID string, quantity int) error {
	r.restored = append(r.restored, domain.OrderItem{ProductID: id, VariantID: variantID, Quantity: quantity})
	return nil
}

type expirationCouponRepository struct {
	domain.CouponRepository
	released []int64
}

func (r *expirationCouponRepository) ReleaseUse(_ context.Context, id int64) error {
	r.released = append(r.released, id)
	return nil
}

type expirationTxManager struct{ repos domain.Repositories }

func (m expirationTxManager) WithinTx(_ context.Context, fn func(domain.Repositories) error) error {
	return fn(m.repos)
}

func TestExpireAwaitingPaymentsReleasesResourcesAtomically(t *testing.T) {
	now := time.Now()
	orderRepo := &expirationOrderRepository{
		order: &domain.Order{
			ID:               7,
			Status:           domain.OrderStatusAwaitingPayment,
			PaymentExpiresAt: now.Add(-time.Second),
			CouponID:         int64Pointer(9),
			Items: []domain.OrderItem{
				{ProductID: 2, Quantity: 3},
				{ProductID: 4, VariantID: "size-m", Quantity: 1},
			},
		},
		candidates: []int64{7},
	}
	paymentRepo := &expirationPaymentRepository{payment: &domain.Payment{
		OrderID: 7, Authority: "authority", Status: domain.PaymentPending,
	}}
	productRepo := &expirationProductRepository{}
	couponRepo := &expirationCouponRepository{}
	service := NewOrderService(orderRepo, nil, expirationTxManager{repos: domain.Repositories{
		Orders: orderRepo, Payments: paymentRepo, Products: productRepo,
		Coupons: couponRepo,
	}}, 30*time.Minute)

	count, err := service.ExpireAwaitingPayments(context.Background(), now, 100)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("expired count = %d, want 1", count)
	}
	if orderRepo.order.Status != domain.OrderStatusCancelled {
		t.Errorf("order status = %q, want cancelled", orderRepo.order.Status)
	}
	if paymentRepo.payment.Status != domain.PaymentCancelled {
		t.Errorf("payment status = %q, want cancelled", paymentRepo.payment.Status)
	}
	if len(productRepo.restored) != 2 || productRepo.restored[1].VariantID != "size-m" {
		t.Errorf("restored stock = %#v, want regular and variant quantities", productRepo.restored)
	}
	if len(couponRepo.released) != 1 || couponRepo.released[0] != 9 {
		t.Errorf("released coupons = %#v, want [9]", couponRepo.released)
	}
}

func TestExpireAwaitingPaymentsRechecksDeadlineAfterLock(t *testing.T) {
	now := time.Now()
	orderRepo := &expirationOrderRepository{
		order: &domain.Order{
			ID:               7,
			Status:           domain.OrderStatusAwaitingPayment,
			PaymentExpiresAt: now.Add(time.Minute),
		},
		candidates: []int64{7},
	}
	productRepo := &expirationProductRepository{}
	service := NewOrderService(orderRepo, nil, expirationTxManager{repos: domain.Repositories{
		Orders: orderRepo, Products: productRepo,
	}}, 30*time.Minute)

	count, err := service.ExpireAwaitingPayments(context.Background(), now, 100)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 || orderRepo.order.Status != domain.OrderStatusAwaitingPayment || len(productRepo.restored) != 0 {
		t.Fatalf("future-expiring order was modified: count=%d status=%q restored=%#v", count, orderRepo.order.Status, productRepo.restored)
	}
}

func int64Pointer(value int64) *int64 { return &value }
