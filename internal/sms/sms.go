// Package sms delivers text messages to phone numbers. ConsoleSender is a
// stand-in for local development and testing; swap in a real Iranian
// provider (Kavenegar, Ghasedak, Melipayamak, ...) later by implementing
// the Sender interface and passing it in cmd/api/main.go instead.
package sms

import (
	"context"
	"fmt"
	"strings"
)

// Sender delivers a text message to a phone number.
type Sender interface {
	Send(ctx context.Context, phone, message string) error
}

// ConsoleSender "sends" messages by printing them to the server's own
// terminal, so OTP codes can be read and tested without a real SMS provider.
type ConsoleSender struct{}

func NewConsoleSender() ConsoleSender { return ConsoleSender{} }

func (ConsoleSender) Send(_ context.Context, phone, message string) error {
	line := strings.Repeat("=", 48)
	fmt.Printf("\n%s\n\U0001F4F1  پیامک شبیه‌سازی‌شده به %s\n%s\n%s\n%s\n\n",
		line, phone, strings.Repeat("-", 48), message, line)
	return nil
}
