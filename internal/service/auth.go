package service

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"shop-api/internal/domain"
)

const otpLength = 6

var otpCodePattern = regexp.MustCompile(`^\d{6}$`)
var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)

// SMSSender delivers the OTP message (implemented by sms.Sender).
type SMSSender interface {
	Send(ctx context.Context, phone, message string) error
}

// TokenIssuer creates access tokens (implemented by auth.TokenManager).
type TokenIssuer interface {
	Generate(user *domain.User) (token string, expiresAt time.Time, err error)
}

// OTPConfig controls how one-time codes behave.
type OTPConfig struct {
	TTL            time.Duration
	ResendCooldown time.Duration
	MaxAttempts    int
}

type LoginResult struct {
	AccessToken string
	ExpiresAt   time.Time
	User        *domain.User
}

// VerifyOTPInput is what's needed to verify a code. FirstName/LastName/Email
// are only required (and only used) the first time a phone number is seen.
type VerifyOTPInput struct {
	Phone     string
	Code      string
	Username  string
	FirstName string
	LastName  string
	Email     string
}

type AuthService struct {
	users  domain.UserRepository
	otps   domain.OTPRepository
	sms    SMSSender
	tokens TokenIssuer
	cfg    OTPConfig
}

func NewAuthService(users domain.UserRepository, otps domain.OTPRepository, sms SMSSender, tokens TokenIssuer, cfg OTPConfig) *AuthService {
	return &AuthService{users: users, otps: otps, sms: sms, tokens: tokens, cfg: cfg}
}

// RequestOTP sends a fresh code to phone (rate-limited) and returns how long
// it stays valid for.
func (s *AuthService) RequestOTP(ctx context.Context, rawPhone string) (time.Duration, error) {
	phone, err := domain.NormalizePhone(rawPhone)
	if err != nil {
		return 0, err
	}

	if existing, err := s.otps.Get(ctx, phone); err == nil {
		wait := s.cfg.ResendCooldown - time.Since(existing.LastSentAt)
		if wait > 0 {
			return 0, &domain.CooldownError{RetryAfterSeconds: int(wait.Seconds()) + 1}
		}
	} else if !errors.Is(err, domain.ErrNotFound) {
		return 0, err
	}

	code, err := randomDigits(otpLength)
	if err != nil {
		return 0, fmt.Errorf("generate one-time code: %w", err)
	}
	if err := s.otps.Upsert(ctx, phone, hashCode(phone, code), time.Now().Add(s.cfg.TTL)); err != nil {
		return 0, err
	}

	message := fmt.Sprintf("کد تایید شما: %s\nاین کد تا %d دقیقه دیگر معتبر است.", code, maxInt(1, int(s.cfg.TTL.Minutes())))
	if err := s.sms.Send(ctx, phone, message); err != nil {
		return 0, fmt.Errorf("send otp sms: %w", err)
	}
	return s.cfg.TTL, nil
}

// VerifyOTP checks the code. If the phone belongs to an existing user, this
// logs them in; otherwise FirstName/LastName are required and a new account
// is created (Email is optional either way).
func (s *AuthService) VerifyOTP(ctx context.Context, in VerifyOTPInput) (*LoginResult, error) {
	phone, err := domain.NormalizePhone(in.Phone)
	if err != nil {
		return nil, err
	}

	code := strings.TrimSpace(in.Code)
	if !otpCodePattern.MatchString(code) {
		return nil, fmt.Errorf("%w: code must be a 6-digit number", domain.ErrInvalidInput)
	}

	row, err := s.otps.Get(ctx, phone)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrOTPNotRequested
		}
		return nil, err
	}
	if time.Now().After(row.ExpiresAt) {
		_ = s.otps.Delete(ctx, phone)
		return nil, domain.ErrOTPExpired
	}
	if row.Attempts >= s.cfg.MaxAttempts {
		return nil, domain.ErrOTPTooManyAttempts
	}
	if subtle.ConstantTimeCompare([]byte(row.CodeHash), []byte(hashCode(phone, code))) != 1 {
		if err := s.otps.IncrementAttempts(ctx, phone); err != nil {
			return nil, fmt.Errorf("record one-time code attempt: %w", err)
		}
		return nil, domain.ErrOTPInvalidCode
	}

	user, err := s.users.GetByPhone(ctx, phone)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}
	if user != nil && !user.IsActive {
		return nil, domain.ErrAccountDisabled
	}

	if user == nil {
		first := strings.TrimSpace(in.FirstName)
		last := strings.TrimSpace(in.LastName)
		fields := map[string]string{}
		if first == "" {
			fields["first_name"] = "is required"
		}
		if last == "" {
			fields["last_name"] = "is required"
		}
		username := strings.ToLower(strings.TrimSpace(in.Username))
		if !usernamePattern.MatchString(username) {
			fields["username"] = "must be 3-30 letters, numbers, or underscores"
		}
		if len(fields) > 0 {
			return nil, domain.NewFieldError(fields)
		}

		var email *string
		if e := strings.TrimSpace(in.Email); e != "" {
			email = &e
		}

		user = &domain.User{FirstName: first, LastName: last, Username: username, Phone: phone, Email: email, Role: domain.RoleUser, IsActive: true}
		if err := s.users.Create(ctx, user); err != nil {
			return nil, err
		}
	}

	if err := s.otps.Delete(ctx, phone); err != nil {
		return nil, fmt.Errorf("consume one-time code: %w", err)
	}

	token, expiresAt, err := s.tokens.Generate(user)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}
	return &LoginResult{AccessToken: token, ExpiresAt: expiresAt, User: user}, nil
}

// EnsureAdmin creates the bootstrap admin account if it does not exist yet.
func (s *AuthService) EnsureAdmin(ctx context.Context, firstName, lastName, rawPhone string) error {
	if rawPhone == "" {
		return nil
	}
	phone, err := domain.NormalizePhone(rawPhone)
	if err != nil {
		return fmt.Errorf("invalid ADMIN_PHONE: %w", err)
	}

	_, err = s.users.GetByPhone(ctx, phone)
	if err == nil {
		return nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return err
	}

	admin := &domain.User{FirstName: firstName, LastName: lastName, Username: "admin_" + strings.TrimLeft(phone, "+0"), Phone: phone, Role: domain.RoleAdmin, IsActive: true}
	if err := s.users.Create(ctx, admin); err != nil && !errors.Is(err, domain.ErrPhoneTaken) {
		return err
	}
	return nil
}

func (s *AuthService) GetProfile(ctx context.Context, userID int64) (*domain.User, error) {
	return s.users.GetByID(ctx, userID)
}

func hashCode(phone, code string) string {
	sum := sha256.Sum256([]byte(phone + ":" + code))
	return hex.EncodeToString(sum[:])
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
