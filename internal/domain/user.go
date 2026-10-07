package domain

import "time"

// Role defines what a user is allowed to do.
type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

// User is an account that can log in (via phone + OTP) and place orders.
type User struct {
	ID        int64
	FirstName string
	LastName  string
	Username  string
	Phone     string
	Email     *string
	Role      Role
	IsActive  bool
	CreatedAt time.Time
}
