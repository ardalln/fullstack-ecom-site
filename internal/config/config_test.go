package config

import (
	"net"
	"strings"
	"testing"
)

func setValidConfigEnv(t *testing.T) {
	t.Helper()
	t.Setenv("JWT_SECRET", strings.Repeat("s", 32))
	t.Setenv("JWT_ACCESS_TTL", "24h")
	t.Setenv("OTP_TTL", "2m")
	t.Setenv("OTP_RESEND_COOLDOWN", "60s")
	t.Setenv("ORDER_PAYMENT_TTL", "30m")
	t.Setenv("OTP_MAX_ATTEMPTS", "5")
	t.Setenv("TRUSTED_PROXIES", "")
}

func TestLoadRejectsNonPositiveLifetimes(t *testing.T) {
	for _, test := range []struct {
		name string
		key  string
	}{
		{name: "JWT", key: "JWT_ACCESS_TTL"},
		{name: "OTP", key: "OTP_TTL"},
		{name: "order payment", key: "ORDER_PAYMENT_TTL"},
	} {
		t.Run(test.name, func(t *testing.T) {
			setValidConfigEnv(t)
			t.Setenv(test.key, "0s")
			if _, err := Load(); err == nil {
				t.Fatalf("Load() succeeded with %s=0s", test.key)
			}
		})
	}
}

func TestLoadRejectsNegativeResendCooldown(t *testing.T) {
	setValidConfigEnv(t)
	t.Setenv("OTP_RESEND_COOLDOWN", "-1s")
	if _, err := Load(); err == nil {
		t.Fatal("Load() succeeded with a negative OTP resend cooldown")
	}
}

func TestParseTrustedProxies(t *testing.T) {
	proxies, err := parseTrustedProxies("10.0.0.1, 192.168.0.0/16, 2001:db8::1")
	if err != nil {
		t.Fatal(err)
	}
	if len(proxies) != 3 {
		t.Fatalf("got %d trusted proxy ranges, want 3", len(proxies))
	}
	for _, test := range []struct {
		index int
		ip    string
	}{
		{index: 0, ip: "10.0.0.1"},
		{index: 1, ip: "192.168.10.20"},
		{index: 2, ip: "2001:db8::1"},
	} {
		if !proxies[test.index].Contains(net.ParseIP(test.ip)) {
			t.Errorf("proxy range %d does not contain %s", test.index, test.ip)
		}
	}
	if _, err := parseTrustedProxies("not-an-ip"); err == nil {
		t.Fatal("parseTrustedProxies accepted an invalid address")
	}
}
