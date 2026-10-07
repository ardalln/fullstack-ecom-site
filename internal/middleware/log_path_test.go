package middleware

import "testing"

func TestSafeLogPathRedactsCapabilityValues(t *testing.T) {
	cases := map[string]string{
		"/api/v1/payments/secret-authority":         "/api/v1/payments/:redacted",
		"/api/v1/payments/secret-authority/confirm": "/api/v1/payments/:redacted/confirm",
		"/api/v1/tracking/KCK-0123456789ABCDEF":     "/api/v1/tracking/:redacted",
		"/api/v1/products/42/comments":              "/api/v1/products/42/comments",
	}
	for input, want := range cases {
		if got := SafeLogPath(input); got != want {
			t.Errorf("SafeLogPath(%q) = %q, want %q", input, got, want)
		}
	}
}
