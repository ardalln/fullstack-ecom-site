package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"shop-api/internal/domain"
)

type OrderService struct {
	orders     domain.OrderRepository
	addresses  domain.AddressRepository
	tx         domain.TxManager
	paymentTTL time.Duration
}

func NewOrderService(orders domain.OrderRepository, addresses domain.AddressRepository, tx domain.TxManager, paymentTTL time.Duration) *OrderService {
	return &OrderService{orders: orders, addresses: addresses, tx: tx, paymentTTL: paymentTTL}
}

// PlaceOrder buys the given items for a user, shipping to one of their saved
// addresses. Everything happens in one database transaction: either the
// stock is decreased AND the order is created, or nothing changes at all.
// The order starts out "awaiting_payment"; RequestPayment (payment.go)
// starts the checkout session.
//
//  1. Look up the shipping address (must belong to the user).
//  2. Lock the product rows (SELECT ... FOR UPDATE, ordered by id).
//  3. Check every product exists and has enough stock.
//  4. Decrease stock and build order items with the price at purchase time.
//  5. Insert the order (with a snapshot of the address) and its items.
func (s *OrderService) PlaceOrder(ctx context.Context, userID, addressID, shippingMethodID int64, couponCode string, items []domain.PurchaseItem) (*domain.Order, error) {
	quantities, productTotals, err := mergeItems(items)
	if err != nil {
		return nil, err
	}

	address, err := s.addresses.GetByID(ctx, userID, addressID)
	if err != nil {
		return nil, err
	}

	productIDs := make([]int64, 0, len(productTotals))
	for id := range productTotals {
		productIDs = append(productIDs, id)
	}
	sort.Slice(productIDs, func(i, j int) bool { return productIDs[i] < productIDs[j] }) // consistent lock order => no deadlocks

	var order *domain.Order
	err = s.tx.WithinTx(ctx, func(repos domain.Repositories) error {
		products, err := repos.Products.GetByIDsForUpdate(ctx, productIDs)
		if err != nil {
			return err
		}
		byID := make(map[int64]domain.Product, len(products))
		for _, p := range products {
			byID[p.ID] = p
		}

		o := &domain.Order{
			UserID:           userID,
			Status:           domain.OrderStatusAwaitingPayment,
			Items:            make([]domain.OrderItem, 0, len(productIDs)),
			PaymentExpiresAt: time.Now().Add(s.paymentTTL),

			ShippingReceiverName:  address.ReceiverName,
			ShippingReceiverPhone: address.ReceiverPhone,
			ShippingProvince:      address.Province,
			ShippingCity:          address.City,
			ShippingAddressLine:   address.AddressLine,
			ShippingPostalCode:    address.PostalCode,
			ShippingMethodName:    "",
		}
		for _, id := range productIDs {
			p, ok := byID[id]
			if !ok {
				return fmt.Errorf("%w: product %d", domain.ErrNotFound, id)
			}
			if !p.IsActive {
				return fmt.Errorf("%w: product %d is not available", domain.ErrNotFound, id)
			}
			if len(p.Variants) == 0 {
				for key := range quantities {
					if key.productID == id && (key.variantID != "" || key.variantName != "") {
						return fmt.Errorf("%w: product %d has no selectable variants", domain.ErrInvalidInput, id)
					}
				}
				qty := productTotals[id]
				if p.Stock < qty {
					return fmt.Errorf("%w: product %d has %d in stock, requested %d", domain.ErrInsufficientStock, id, p.Stock, qty)
				}
				if err := repos.Products.DecreaseStock(ctx, id, qty); err != nil {
					return err
				}
				unitPrice := p.Price
				if p.DiscountPrice > 0 {
					unitPrice = p.DiscountPrice
				}
				if err := addOrderAmount(o, unitPrice, qty); err != nil {
					return err
				}
				o.Items = append(o.Items, domain.OrderItem{ProductID: id, ProductName: p.Name, Quantity: qty, UnitPrice: unitPrice})
				continue
			}

			lineQuantities := make(map[string]int)
			for key, qty := range quantities {
				if key.productID != id {
					continue
				}
				selected := -1
				for index, variant := range p.Variants {
					if key.variantID != "" && key.variantID == variant.ID {
						selected = index
						break
					}
					if key.variantID == "" && key.variantName != "" && strings.EqualFold(strings.TrimSpace(variant.Name), key.variantName) {
						selected = index
						break
					}
				}
				if selected < 0 {
					return fmt.Errorf("%w: selected variant is unavailable for product %d", domain.ErrInvalidInput, id)
				}
				variant := p.Variants[selected]
				lineQuantities[variant.ID] += qty
			}
			variantIDs := make([]string, 0, len(lineQuantities))
			for variantID := range lineQuantities {
				variantIDs = append(variantIDs, variantID)
			}
			sort.Strings(variantIDs)
			for _, variantID := range variantIDs {
				lineQty := lineQuantities[variantID]
				var selected domain.ProductVariant
				for _, variant := range p.Variants {
					if variant.ID == variantID {
						selected = variant
						break
					}
				}
				if selected.Stock < lineQty {
					return fmt.Errorf("%w: variant %s has %d in stock, requested %d", domain.ErrInsufficientStock, selected.Name, selected.Stock, lineQty)
				}
				if err := repos.Products.DecreaseVariantStock(ctx, id, variantID, lineQty); err != nil {
					return err
				}
				unitPrice := selected.Price
				if selected.DiscountPrice > 0 {
					unitPrice = selected.DiscountPrice
				}
				if err := addOrderAmount(o, unitPrice, lineQty); err != nil {
					return err
				}
				o.Items = append(o.Items, domain.OrderItem{ProductID: id, ProductName: p.Name, VariantID: variantID,
					VariantLabel: p.VariantLabel, VariantName: selected.Name, Quantity: lineQty, UnitPrice: unitPrice})
			}
		}

		method, err := repos.Shipping.GetByID(ctx, shippingMethodID)
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrShippingUnavailable
		}
		if err != nil {
			return err
		}
		if !method.IsActive || (method.Province != "" && !strings.EqualFold(method.Province, address.Province)) || (method.City != "" && !strings.EqualFold(method.City, address.City)) {
			return domain.ErrShippingUnavailable
		}
		methodID := method.ID
		o.ShippingMethodID = &methodID
		o.ShippingMethodName = method.Name
		o.ShippingCost = method.Price
		couponCode = strings.TrimSpace(couponCode)
		if couponCode != "" {
			coupon, err := repos.Coupons.GetByCodeForUpdate(ctx, couponCode)
			if errors.Is(err, domain.ErrNotFound) {
				return domain.ErrCouponInvalid
			}
			if err != nil {
				return err
			}
			now := time.Now()
			if !coupon.IsActive || (coupon.StartsAt != nil && now.Before(*coupon.StartsAt)) || (coupon.EndsAt != nil && !now.Before(*coupon.EndsAt)) || (coupon.UsageLimit != nil && coupon.UsedCount >= *coupon.UsageLimit) {
				return domain.ErrCouponInvalid
			}
			discount, err := couponDiscount(coupon, o.ItemsAmount)
			if err != nil {
				return err
			}
			if err := repos.Coupons.ReserveUse(ctx, coupon.ID, now); err != nil {
				return err
			}
			o.CouponCode = coupon.Code
			couponID := coupon.ID
			o.CouponID = &couponID
			o.DiscountAmount = discount
		}
		if o.ItemsAmount-o.DiscountAmount > int64(^uint64(0)>>1)-o.ShippingCost {
			return fmt.Errorf("%w: order amount is too large", domain.ErrInvalidInput)
		}
		o.TotalAmount = o.ItemsAmount - o.DiscountAmount + o.ShippingCost
		trackingToken, err := randomToken(8)
		if err != nil {
			return fmt.Errorf("generate tracking code: %w", err)
		}
		o.TrackingCode = "KCK-" + strings.ToUpper(trackingToken)

		if err := repos.Orders.Create(ctx, o); err != nil {
			return err
		}
		order = o
		return nil
	})
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (s *OrderService) ExpireAwaitingPayments(ctx context.Context, now time.Time, limit int) (int, error) {
	ids, err := s.orders.ListAwaitingPaymentBefore(ctx, now, limit)
	if err != nil {
		return 0, err
	}
	expired := 0
	var failures []error
	for _, id := range ids {
		didExpire := false
		err := s.tx.WithinTx(ctx, func(repos domain.Repositories) error {
			order, err := repos.Orders.GetByIDForUpdate(ctx, id)
			if err != nil {
				return err
			}
			if order.Status != domain.OrderStatusAwaitingPayment || order.PaymentExpiresAt.After(now) {
				return nil
			}

			payment, err := repos.Payments.GetByOrderID(ctx, id)
			if err == nil {
				if payment.Status != domain.PaymentPending {
					return domain.ErrPaymentAlreadyProcessed
				}
				if err := repos.Payments.MarkCancelled(ctx, payment.Authority); err != nil {
					return err
				}
			} else if !errors.Is(err, domain.ErrNotFound) {
				return err
			}

			if err := repos.Orders.TransitionStatus(ctx, id, domain.OrderStatusAwaitingPayment, domain.OrderStatusCancelled); err != nil {
				return err
			}
			for _, item := range order.Items {
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
			didExpire = true
			return nil
		})
		if err != nil {
			failures = append(failures, fmt.Errorf("expire order %d: %w", id, err))
			continue
		}
		if didExpire {
			expired++
		}
	}
	return expired, errors.Join(failures...)
}

// Get returns an order only if it belongs to the given user. Other users'
// orders look like "not found" so their existence is not leaked.
func (s *OrderService) Get(ctx context.Context, userID, orderID int64) (*domain.Order, error) {
	order, err := s.orders.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.UserID != userID {
		return nil, domain.ErrNotFound
	}
	return order, nil
}

func (s *OrderService) ListByUser(ctx context.Context, userID int64, page, limit int) ([]domain.Order, int, error) {
	return s.orders.ListByUserID(ctx, userID, limit, (page-1)*limit)
}

type orderLineKey struct {
	productID   int64
	variantID   string
	variantName string
}

func addOrderAmount(order *domain.Order, unitPrice int64, quantity int) error {
	maxInt64 := int64(^uint64(0) >> 1)
	if unitPrice <= 0 || unitPrice > maxInt64/int64(quantity) {
		return fmt.Errorf("%w: order amount is too large", domain.ErrInvalidInput)
	}
	subtotal := unitPrice * int64(quantity)
	if order.ItemsAmount > maxInt64-subtotal {
		return fmt.Errorf("%w: order amount is too large", domain.ErrInvalidInput)
	}
	order.ItemsAmount += subtotal
	return nil
}

// mergeItems validates the request and sums quantities of duplicate variants.
func mergeItems(items []domain.PurchaseItem) (map[orderLineKey]int, map[int64]int, error) {
	if len(items) == 0 {
		return nil, nil, fmt.Errorf("%w: order must contain at least one item", domain.ErrInvalidInput)
	}
	merged := make(map[orderLineKey]int, len(items))
	totals := make(map[int64]int, len(items))
	for _, it := range items {
		if it.ProductID <= 0 || it.Quantity <= 0 {
			return nil, nil, fmt.Errorf("%w: product_id and quantity must be positive", domain.ErrInvalidInput)
		}
		variantName := strings.TrimSpace(it.VariantName)
		variantID := strings.TrimSpace(it.VariantID)
		if len(variantID) > 80 {
			return nil, nil, fmt.Errorf("%w: variant_id must be at most 80 characters", domain.ErrInvalidInput)
		}
		if len([]rune(variantName)) > 80 {
			return nil, nil, fmt.Errorf("%w: variant_name must be at most 80 characters", domain.ErrInvalidInput)
		}
		merged[orderLineKey{productID: it.ProductID, variantID: variantID, variantName: variantName}] += it.Quantity
		totals[it.ProductID] += it.Quantity
	}
	return merged, totals, nil
}
