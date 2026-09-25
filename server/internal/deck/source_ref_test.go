package deck

import (
	"errors"
	"testing"
)

// ParseDeckURL is ADR 0095's deck key: the identity a coverage report
// is cached under and a deck request is deduplicated on. Two spellings
// of one link must give one key, and nothing but a supported host may
// give one at all.
func TestParseDeckURL(t *testing.T) {
	for _, tc := range []struct {
		in, key, url string
	}{
		{"https://moxfield.com/decks/AbC123", "moxfield:AbC123", "https://moxfield.com/decks/AbC123"},
		{"https://www.moxfield.com/decks/AbC123/my-slug?tab=stats", "moxfield:AbC123", "https://moxfield.com/decks/AbC123"},
		{"  https://api2.moxfield.com/decks/AbC123  ", "moxfield:AbC123", "https://moxfield.com/decks/AbC123"},
		{"https://archidekt.com/decks/123456/elves", "archidekt:123456", "https://archidekt.com/decks/123456"},
		{"http://www.archidekt.com/decks/123456", "archidekt:123456", "https://archidekt.com/decks/123456"},
	} {
		ref, err := ParseDeckURL(tc.in)
		if err != nil {
			t.Errorf("%q: %v", tc.in, err)
			continue
		}
		if ref.Key() != tc.key || ref.URL() != tc.url {
			t.Errorf("%q: key %q url %q, want %q %q", tc.in, ref.Key(), ref.URL(), tc.key, tc.url)
		}
	}

	for _, bad := range []string{
		"",
		"https://example.com/decks/1",
		"ftp://moxfield.com/decks/1",
		"https://moxfield.com/users/someone",
		"https://archidekt.com/decks/",
		"https://moxfield.com.evil.example/decks/1",
	} {
		if _, err := ParseDeckURL(bad); !errors.Is(err, ErrUnknownSource) {
			t.Errorf("%q: err = %v, want ErrUnknownSource", bad, err)
		}
	}
}
