package service

import (
	"strings"
	"unicode"
)

// makeEntitySlug creates stable, URL-friendly slugs for categories and brands.
// An empty result is rejected by the caller so names that cannot be transliterated
// can still be saved by supplying a Latin slug explicitly.
func makeEntitySlug(raw, name string) string {
	if strings.TrimSpace(raw) == "" {
		raw = name
	}
	var b strings.Builder
	separator := false
	for _, r := range strings.ToLower(strings.TrimSpace(raw)) {
		if replacement, ok := persianSlug[r]; ok {
			b.WriteString(replacement)
			separator = false
			continue
		}
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			separator = false
			continue
		}
		if r == '-' || r == '_' || unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsSymbol(r) {
			if b.Len() > 0 && !separator {
				b.WriteByte('-')
				separator = true
			}
		}
	}
	slug := strings.Trim(b.String(), "-")
	if len(slug) > 180 {
		slug = strings.TrimRight(slug[:180], "-")
	}
	return slug
}
