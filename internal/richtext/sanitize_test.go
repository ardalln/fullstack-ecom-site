package richtext

import (
	"strings"
	"testing"
)

func TestSanitizeKeepsFormattingAndRemovesUnsafeMarkup(t *testing.T) {
	input := `<h2 onclick="run()">راهنمای کتانی</h2><script>alert(1)</script><p><strong>متن مهم</strong> <a href="javascript:alert(1)">لینک بد</a> <a href="https://example.com" target="_blank">لینک امن</a></p>`
	got := Sanitize(input)
	for _, want := range []string{"<h2>راهنمای کتانی</h2>", "<strong>متن مهم</strong>", `href="https://example.com"`, `rel="noopener noreferrer"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("sanitized markup %q does not contain %q", got, want)
		}
	}
	for _, forbidden := range []string{"onclick", "script", "javascript:"} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("sanitized markup still contains %q: %s", forbidden, got)
		}
	}
}

func TestPlainTextRejectsEmptyFormattedContent(t *testing.T) {
	if got := PlainText(`<p><br></p><script>alert(1)</script>`); got != "" {
		t.Fatalf("PlainText() = %q, want empty", got)
	}
}

func TestSanitizePreservesEditorBlocksAndInlineFormatting(t *testing.T) {
	input := `<div><h3>راهنما</h3><p><s>قدیمی</s> <sub>۲</sub> <sup>۳</sup></p><hr><p><a href="tel:+982100000000">تماس</a></p></div>`
	got := Sanitize(input)
	for _, want := range []string{"<div>", "<h3>راهنما</h3>", "<s>قدیمی</s>", "<sub>۲</sub>", "<sup>۳</sup>", "<hr/>", `href="tel:+982100000000"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("sanitized markup %q does not contain %q", got, want)
		}
	}
}
