package domain

import "time"

// OrderStatus is the lifecycle state of an order.
type OrderStatus string

const (
	OrderStatusAwaitingPayment OrderStatus = "awaiting_payment"
	OrderStatusPaid            OrderStatus = "paid"
	OrderStatusProcessing      OrderStatus = "processing"
	OrderStatusShipped         OrderStatus = "shipped"
	OrderStatusCancelled       OrderStatus = "cancelled"
)

// PurchaseItem is one line of a purchase request.
type PurchaseItem struct {
	ProductID   int64
	VariantID   string
	VariantName string
	Quantity    int
}

// OrderItem is a snapshot of a product at the time of purchase.
type OrderItem struct {
	ID           int64
	OrderID      int64
	ProductID    int64
	ProductName  string
	VariantID    string
	VariantLabel string
	VariantName  string
	Quantity     int
	UnitPrice    int64
}

// OrderPaymentInfo is a lightweight view of the order's payment (if one has
// been started yet), embedded in Order for display purposes.
type OrderPaymentInfo struct {
	Authority string
	Status    PaymentStatus
	RefID     *string
	PaidAt    *time.Time
}

// Order is a purchase made by a user. The shipping fields are a snapshot of
// the address chosen at checkout, so later edits or deletion of the saved
// address never change what a past order says it was shipped to.
type Order struct {
	ID                    int64
	UserID                int64
	CustomerName          string
	CustomerPhone         string
	Status                OrderStatus
	TotalAmount           int64
	ItemsAmount           int64
	DiscountAmount        int64
	ShippingCost          int64
	CouponID              *int64
	CouponCode            string
	ShippingMethodID      *int64
	ShippingMethodName    string
	TrackingCode          string
	ShippingReceiverName  string
	ShippingReceiverPhone string
	ShippingProvince      string
	ShippingCity          string
	ShippingAddressLine   string
	ShippingPostalCode    string
	Items                 []OrderItem
	Payment               *OrderPaymentInfo
	PaymentExpiresAt      time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
}
