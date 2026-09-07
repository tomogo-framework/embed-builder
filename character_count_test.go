package embedbuilder

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestCharacterCountBoundaries(t *testing.T) {
	t.Parallel()

	for prefix := 0; prefix < 32; prefix++ {
		for suffix := 0; suffix < 16; suffix++ {
			for _, middle := range []string{"", "é", "🙂", "\xff", "\xc3", "\xed\xa0\x80"} {
				value := "\u2003\t" + strings.Repeat("a", prefix) + middle + strings.Repeat("b", suffix) + " \u2003"
				want := utf8.RuneCountInString(strings.TrimSpace(value))
				if got := characterCount(value); got != want {
					t.Fatalf("characterCount(%q) = %d, want %d", value, got, want)
				}
			}
		}
	}
}

func FuzzCharacterCount(f *testing.F) {
	for _, value := range []string{"", "ASCII", "\t\u2003 ", "1234567🙂", "12345678\xff", "é🙂"} {
		f.Add(value)
	}

	f.Fuzz(func(t *testing.T, value string) {
		want := utf8.RuneCountInString(strings.TrimSpace(value))
		if got := characterCount(value); got != want {
			t.Fatalf("characterCount(%q) = %d, want %d", value, got, want)
		}
	})
}

func BenchmarkCharacterCount(b *testing.B) {
	for _, test := range []struct {
		name  string
		value string
	}{
		{
			name:  "ShortASCII",
			value: "field name",
		},
		{
			name:  "LongASCII",
			value: strings.Repeat("x", 3000),
		},
		{
			name:  "Unicode",
			value: strings.Repeat("🙂", 1000),
		},
		{
			name:  "Mixed",
			value: strings.Repeat("x", 2999) + "🙂",
		},
	} {
		b.Run(test.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = characterCount(test.value)
			}
		})
	}
}
