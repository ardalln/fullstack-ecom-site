package router

import (
	"net"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestTrustedProxyIPExtraction(t *testing.T) {
	_, trustedRange, err := net.ParseCIDR("10.0.0.0/8")
	if err != nil {
		t.Fatal(err)
	}
	extractIP := echo.ExtractIPFromXFFHeader(echo.TrustIPRange(trustedRange))

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.2:8080"
	req.Header.Set(echo.HeaderXForwardedFor, "198.51.100.25, 10.0.0.1")
	if got := extractIP(req); got != "198.51.100.25" {
		t.Errorf("trusted proxy IP = %q, want 198.51.100.25", got)
	}

	req.RemoteAddr = "203.0.113.8:8080"
	req.Header.Set(echo.HeaderXForwardedFor, "198.51.100.25")
	if got := extractIP(req); got != "203.0.113.8" {
		t.Errorf("untrusted peer IP = %q, want 203.0.113.8", got)
	}
}
