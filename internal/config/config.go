package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all runtime configuration, loaded from environment variables.
type Config struct {
	AppPort      string
	DatabaseURL  string
	JWTSecret    string
	JWTIssuer    string
	JWTAccessTTL time.Duration

	OTPTTL            time.Duration
	OTPResendCooldown time.Duration
	OTPMaxAttempts    int

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

	otpTTL, err := time.ParseDuration(getEnv("OTP_TTL", "2m"))
	if err != nil {
		return nil, fmt.Errorf("invalid OTP_TTL: %w", err)
	}

	otpCooldown, err := time.ParseDuration(getEnv("OTP_RESEND_COOLDOWN", "60s"))
	if err != nil {
		return nil, fmt.Errorf("invalid OTP_RESEND_COOLDOWN: %w", err)
	}

	otpMaxAttempts, err := strconv.Atoi(getEnv("OTP_MAX_ATTEMPTS", "5"))
	if err != nil || otpMaxAttempts < 1 {
		return nil, errors.New("OTP_MAX_ATTEMPTS must be a positive integer")
	}

	cfg := &Config{
		AppPort:           getEnv("APP_PORT", "8080"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/shop?sslmode=disable"),
		JWTSecret:         os.Getenv("JWT_SECRET"),
		JWTIssuer:         getEnv("JWT_ISSUER", "shop-api"),
		JWTAccessTTL:      ttl,
		OTPTTL:            otpTTL,
		OTPResendCooldown: otpCooldown,
		OTPMaxAttempts:    otpMaxAttempts,
		AdminFirstName:    getEnv("ADMIN_FIRST_NAME", "مدیر"),
		AdminLastName:     getEnv("ADMIN_LAST_NAME", "فروشگاه"),
		AdminPhone:        os.Getenv("ADMIN_PHONE"),
	}

	if len(cfg.JWTSecret) < 32 {
		return nil, errors.New("JWT_SECRET must be set and at least 32 characters long")
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
