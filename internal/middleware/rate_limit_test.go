package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

type testRateLimitStore struct {
	allowed bool
	err     error
	key     string
}

func (s *testRateLimitStore) Allow(_ context.Context, key string, _ int, _ time.Duration) (bool, error) {
	s.key = key
	return s.allowed, s.err
}

func TestSharedRateLimitRejectsRequests(t *testing.T) {
	store := &testRateLimitStore{}
	e := echo.New()
	e.POST("/api/v1/auth/otp/request", func(c echo.Context) error {
		t.Fatal("rate limited request reached its handler")
		return nil
	}, SharedRateLimit(store, 2, 30*time.Second))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/otp/request", nil)
	req.RemoteAddr = "192.0.2.5:1234"
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}
	if got := rec.Header().Get("Retry-After"); got != "30" {
		t.Errorf("Retry-After = %q, want 30", got)
	}
	if store.key == "" || len(store.key) != 64 {
		t.Errorf("rate limit key was not hashed: %q", store.key)
	}
}

func TestSharedRateLimitReturnsStoreErrors(t *testing.T) {
	store := &testRateLimitStore{err: errors.New("database unavailable")}
	e := echo.New()
	e.POST("/limited", func(c echo.Context) error { return c.NoContent(http.StatusNoContent) }, SharedRateLimit(store, 1, time.Second))

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/limited", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}
