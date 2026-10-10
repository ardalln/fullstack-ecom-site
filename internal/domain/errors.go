package domain

import "errors"

// Sentinel errors. The HTTP layer maps them to status codes.
var (
	ErrInvalidInput      = errors.New("invalid input")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("you do not have permission to perform this action")
	ErrNotFound          = errors.New("resource not found")
	ErrRateLimitExceeded = errors.New("rate limit exceeded")

	ErrPhoneTaken        = errors.New("phone number is already registered")
	ErrEmailTaken        = errors.New("email is already registered")
	ErrCategoryNameTaken = errors.New("category name is already in use")
	ErrCategorySlugTaken = errors.New("category slug is already in use")
	ErrCategoryInUse     = errors.New("category is assigned to one or more products")
	ErrBrandNameTaken    = errors.New("brand name is already in use")
	ErrBrandSlugTaken    = errors.New("brand slug is already in use")
	ErrProductSlugTaken  = errors.New("product slug is already in use")
	ErrBrandInUse        = errors.New("brand is assigned to one or more products")
	ErrBlogSlugTaken     = errors.New("blog post slug is already in use")
	ErrTicketClosed      = errors.New("ticket is closed")

	ErrOTPNotRequested    = errors.New("no verification code was requested for this phone number")
	ErrOTPExpired         = errors.New("verification code has expired")
	ErrOTPInvalidCode     = errors.New("verification code is incorrect")
	ErrOTPTooManyAttempts = errors.New("too many incorrect attempts, request a new code")

	ErrInsufficientStock   = errors.New("insufficient stock")
	ErrProductInUse        = errors.New("product is referenced by existing orders and cannot be deleted")
	ErrProductVariantInUse = errors.New("product variant is used by a pending order and cannot be removed")

	ErrOrderNotPayable         = errors.New("order is not awaiting payment")
	ErrPaymentAlreadyProcessed = errors.New("payment has already been processed")
	ErrOrderStatusConflict     = errors.New("order status cannot be changed to the requested state")
	ErrLastAdmin               = errors.New("at least one active admin account must remain")
	ErrAccountDisabled         = errors.New("this account is disabled")
	ErrUsernameTaken           = errors.New("username is already in use")
	ErrCouponInvalid           = errors.New("coupon code is invalid or no longer available")
	ErrCouponMinimum           = errors.New("order total does not meet the coupon minimum")
	ErrShippingUnavailable     = errors.New("shipping method is unavailable for this destination")
)

// FieldError carries per-field validation messages that were not caught by
// struct-tag validation (e.g. "first name is required" only for new users).
// The HTTP layer maps it to the same 422 response shape as tag validation.
type FieldError struct{ Fields map[string]string }

func (e *FieldError) Error() string { return "validation failed" }

func NewFieldError(fields map[string]string) error { return &FieldError{Fields: fields} }

// CooldownError signals the caller must wait before retrying a request
// (used for OTP resend rate limiting).
type CooldownError struct{ RetryAfterSeconds int }

func (e *CooldownError) Error() string { return "please wait before requesting another code" }
