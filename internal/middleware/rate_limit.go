package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"shop-api/internal/domain"
)

type rateLimitStore interface {
	Allow(ctx context.Context, key string, capacity int, refillInterval time.Duration) (bool, error)
}

func SharedRateLimit(store rateLimitStore, capacity int, refillInterval time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := c.Request().Method + "\x00" + c.Path() + "\x00" + c.RealIP()
			sum := sha256.Sum256([]byte(key))
			allowed, err := store.Allow(c.Request().Context(), hex.EncodeToString(sum[:]), capacity, refillInterval)
			if err != nil {
				return fmt.Errorf("apply shared rate limit: %w", err)
			}
			if !allowed {
				retryAfter := int(refillInterval.Seconds())
				if retryAfter < 1 {
					retryAfter = 1
				}
				c.Response().Header().Set(echo.HeaderRetryAfter, strconv.Itoa(retryAfter))
				return echo.NewHTTPError(http.StatusTooManyRequests, domain.ErrRateLimitExceeded.Error())
			}
			return next(c)
		}
	}
}
