package domain

import "time"

// Address is a saved shipping address a user can pick at checkout.
type Address struct {
	ID            int64
	UserID        int64
	ReceiverName  string
	ReceiverPhone string
	Province      string
	City          string
	AddressLine   string
	PostalCode    string
	IsDefault     bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
