package domain

import "time"

// PaymentStatus is the lifecycle state of a payment session.
type PaymentStatus string

const (
	PaymentPending   PaymentStatus = "pending"
	PaymentPaid      PaymentStatus = "paid"
	PaymentCancelled PaymentStatus = "cancelled"
)

// Payment is one checkout session for an order, modeled after how a real
// Iranian gateway (e.g. ZarinPal) works: an "authority" token identifies the
// session, and a "ref id" (tracking code) is issued once it is paid.
type Payment struct {
	ID        int64
	OrderID   int64
	Authority string
	Amount    int64
	Status    PaymentStatus
	RefID     *string
	PaidAt    *time.Time
	CreatedAt time.Time
}
