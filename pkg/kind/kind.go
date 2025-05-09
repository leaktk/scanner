package kind

import (
	"strings"
	"unicode"
)

// KindsMatch normalizes kinds for easier matches
func KindsMatch(a, b string) bool {
	return NormalizeKind(a) == NormalizeKind(b)
}

// NormalizeKind lower cases and removes dashes and underscores to allow some
// flexibility when comparing kinds
func NormalizeKind(kind string) string {
	var b strings.Builder
	b.Grow(len(kind))

	for _, r := range kind {
		r := unicode.ToLower(r)
		if r != '_' && r != '-' {
			b.WriteRune(r)
		}
	}

	return b.String()
}
