package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all runtime configuration, loaded from environment variables.
type Config struct {
	AppPort         string
	DatabaseURL     string
	JWTSecret       string
	JWTIssuer       string
	JWTAccessTTL    time.Duration
	OrderPaymentTTL time.Duration

	OTPTTL            time.Duration
	OTPResendCooldown time.Duration
	OTPMaxAttempts    int
	TrustedProxies    []net.IPNet

	// Optional bootstrap admin account, created on startup if it doesn't
	// already exist. Leave ADMIN_PHONE empty to skip.
	AdminFirstName string
	AdminLastName  string
	AdminPhone     string
}

// Load reads and validates the configuration.
func Load() (*Config, error) {
	ttl, err := time.ParseDuration(getEnv("JWT_ACCESS_TTL", "24h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_TTL: %w", err)
	}
	if ttl <= 0 {
		return nil, errors.New("JWT_ACCESS_TTL must be positive")
	}

	orderPaymentTTL, err := time.ParseDuration(getEnv("ORDER_PAYMENT_TTL", "30m"))
	if err != nil {
		return nil, fmt.Errorf("invalid ORDER_PAYMENT_TTL: %w", err)
	}
	if orderPaymentTTL <= 0 {
		return nil, errors.New("ORDER_PAYMENT_TTL must be positive")
	}

	otpTTL, err := time.ParseDuration(getEnv("OTP_TTL", "2m"))
	if err != nil {
		return nil, fmt.Errorf("invalid OTP_TTL: %w", err)
	}
	if otpTTL <= 0 {
		return nil, errors.New("OTP_TTL must be positive")
	}

	otpCooldown, err := time.ParseDuration(getEnv("OTP_RESEND_COOLDOWN", "60s"))
	if err != nil {
		return nil, fmt.Errorf("invalid OTP_RESEND_COOLDOWN: %w", err)
	}
	if otpCooldown < 0 {
		return nil, errors.New("OTP_RESEND_COOLDOWN must not be negative")
	}

	otpMaxAttempts, err := strconv.Atoi(getEnv("OTP_MAX_ATTEMPTS", "5"))
	if err != nil || otpMaxAttempts < 1 {
		return nil, errors.New("OTP_MAX_ATTEMPTS must be a positive integer")
	}
	trustedProxies, err := parseTrustedProxies(os.Getenv("TRUSTED_PROXIES"))
	if err != nil {
		return nil, fmt.Errorf("invalid TRUSTED_PROXIES: %w", err)
	}

	cfg := &Config{
		AppPort:           getEnv("APP_PORT", "8080"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/shop?sslmode=disable"),
		JWTSecret:         os.Getenv("JWT_SECRET"),
		JWTIssuer:         getEnv("JWT_ISSUER", "shop-api"),
		JWTAccessTTL:      ttl,
		OrderPaymentTTL:   orderPaymentTTL,
		OTPTTL:            otpTTL,
		OTPResendCooldown: otpCooldown,
		OTPMaxAttempts:    otpMaxAttempts,
		TrustedProxies:    trustedProxies,
		AdminFirstName:    getEnv("ADMIN_FIRST_NAME", "مدیر"),
		AdminLastName:     getEnv("ADMIN_LAST_NAME", "فروشگاه"),
		AdminPhone:        os.Getenv("ADMIN_PHONE"),
	}

	if len(cfg.JWTSecret) < 32 {
		return nil, errors.New("JWT_SECRET must be set and at least 32 characters long")
	}
	return cfg, nil
}

func parseTrustedProxies(raw string) ([]net.IPNet, error) {
	var proxies []net.IPNet
	for _, value := range strings.Split(raw, ",") {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, network, err := net.ParseCIDR(value); err == nil {
			proxies = append(proxies, *network)
			continue
		}
		ip := net.ParseIP(value)
		if ip == nil {
			return nil, fmt.Errorf("%q is not an IP address or CIDR", value)
		}
		if ipv4 := ip.To4(); ipv4 != nil {
			proxies = append(proxies, net.IPNet{IP: ipv4, Mask: net.CIDRMask(32, 32)})
		} else {
			proxies = append(proxies, net.IPNet{IP: ip, Mask: net.CIDRMask(128, 128)})
		}
	}
	return proxies, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
