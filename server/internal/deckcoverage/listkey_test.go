package deckcoverage

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
)

func listKeyOf(t *testing.T, idx *cards.Index, text string) string {
	t.Helper()
	entries, err := deck.ParseText(text)
	if err != nil {
		t.Fatalf("ParseText: %v", err)
	}
	key, err := ListKey(idx, entries)
	if err != nil {
		t.Fatalf("ListKey: %v", err)
	}
	return key
}

// Two exports of one deck — section headers against a *CMDR* marker,
// set codes and collector numbers against bare names, one row against
// two, different order and case — are one deck.
func TestListKeyIgnoresFormatting(t *testing.T) {
	idx, tc := TestingIndex(t)
	sectioned := fmt.Sprintf(`Commander
1 %s (TST) 1 *F*

Deck
1 %s (TST) 7
30 %s (TST) 280
2 %s (ABC) 281
1 %s
1 Some Card Nobody Printed
SB: 1 %s
`, tc.Commander, tc.Manual, tc.Basic, tc.Basic, tc.Automated, tc.Unreviewed)
	flat := fmt.Sprintf(`1x %s
32 %s
1 %s
1 some card   nobody printed
1 %s *CMDR*
`, strings.ToLower(tc.Automated), tc.Basic, tc.Manual, tc.Commander)
	a, b := listKeyOf(t, idx, sectioned), listKeyOf(t, idx, flat)
	if a != b {
		t.Errorf("two exports of one deck: %s != %s", a, b)
	}
	if !strings.HasPrefix(a, ListKeyPrefix) || len(a) != len(ListKeyPrefix)+16 || !IsListKey(a) {
		t.Errorf("key = %q, want list:<16 hex>", a)
	}
}

// The same cards under another commander are another deck, and so is
// the same list with one card more.
func TestListKeyDistinguishesDecks(t *testing.T) {
	idx, tc := TestingIndex(t)
	base := listKeyOf(t, idx, fmt.Sprintf("1 %s *CMDR*\n1 %s\n1 %s\n", tc.Commander, tc.Vanilla, tc.Manual))
	for name, other := range map[string]string{
		"another commander": fmt.Sprintf("1 %s *CMDR*\n1 %s\n1 %s\n", tc.Vanilla, tc.Commander, tc.Manual),
		"no commander":      fmt.Sprintf("1 %s\n1 %s\n1 %s\n", tc.Commander, tc.Vanilla, tc.Manual),
		"one more copy":     fmt.Sprintf("1 %s *CMDR*\n1 %s\n2 %s\n", tc.Commander, tc.Vanilla, tc.Manual),
	} {
		if k := listKeyOf(t, idx, other); k == base {
			t.Errorf("%s: same key %s", name, k)
		}
	}
}

func TestListKeyRefusesAHugeList(t *testing.T) {
	idx, tc := TestingIndex(t)
	entries := []deck.Entry{{Name: tc.Basic, Count: MaxListCopies + 1}}
	if _, err := ListKey(idx, entries); !errors.Is(err, ErrListTooLong) {
		t.Errorf("err = %v, want ErrListTooLong", err)
	}
	if _, err := ListKey(nil, entries); !errors.Is(err, ErrNoIndex) {
		t.Errorf("nil index: err = %v", err)
	}
}
