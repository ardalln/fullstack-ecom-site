package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
)

// randomDigits returns a cryptographically random numeric string of length n
// (e.g. randomDigits(6) -> "042917"), used for OTP codes and tracking numbers.
func randomDigits(n int) (string, error) {
	max := big.NewInt(1)
	ten := big.NewInt(10)
	for i := 0; i < n; i++ {
		max.Mul(max, ten)
	}
	v, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%0*d", n, v.Int64()), nil
}

// randomToken returns a random hex string built from n random bytes, used
// for unguessable payment authority tokens.
func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
