package domain

import (
	"fmt"
	"regexp"
	"strings"
)

var iranMobilePattern = regexp.MustCompile(`^09\d{9}$`)

// NormalizePhone accepts the common ways an Iranian mobile number might be
// typed (+98, 0098, 98, or without the leading 0) and returns it in the
// canonical 09xxxxxxxxx form used everywhere in this app.
func NormalizePhone(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	s = strings.NewReplacer(" ", "", "-", "", "(", "", ")", "").Replace(s)

	switch {
	case strings.HasPrefix(s, "+98"):
		s = "0" + strings.TrimPrefix(s, "+98")
	case strings.HasPrefix(s, "0098"):
		s = "0" + strings.TrimPrefix(s, "0098")
	case strings.HasPrefix(s, "98") && len(s) == 12:
		s = "0" + strings.TrimPrefix(s, "98")
	case strings.HasPrefix(s, "9") && len(s) == 10:
		s = "0" + s
	}

	if !iranMobilePattern.MatchString(s) {
		return "", fmt.Errorf("%w: phone must be a valid Iranian mobile number, e.g. 0912xxxxxxx", ErrInvalidInput)
	}
	return s, nil
}
