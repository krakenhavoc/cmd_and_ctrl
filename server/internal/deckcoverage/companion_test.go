package deckcoverage

import (
	"reflect"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
)

// ADR 0172 §2: the companion is bucketed on its own and kept out of the
// deck's cards and counts, so the copies still total the deck's size.
func TestCompanionIsReportedOutsideTheDeck(t *testing.T) {
	idx, tc := TestingIndex(t)
	entries := append(testingDeck(tc), deck.Entry{Name: tc.Automated, Count: 1, IsCompanion: true})
	r, err := Build(idx, Deck{Source: "text", Entries: entries})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if r.Companion == nil || r.Companion.Name != tc.Automated || r.Companion.Bucket != Automated || r.Companion.Count != 1 {
		t.Fatalf("companion = %+v", r.Companion)
	}
	// The same copies as the deck without it.
	wantCopies := map[Bucket]int{Manual: 1, Unreviewed: 1, Caveats: 1, Automated: 1, NoEffect: 34}
	if !reflect.DeepEqual(r.Copies, wantCopies) {
		t.Errorf("copies = %v, want %v", r.Copies, wantCopies)
	}
	for _, c := range r.Cards {
		if c.Name == tc.Automated && c.Count != 1 {
			t.Errorf("the companion was counted into the deck's %s row: %d", c.Name, c.Count)
		}
	}

	// An unknown companion is in no bucket and not part of the size.
	entries = append(testingDeck(tc), deck.Entry{Name: "No Such Companion", Count: 1, IsCompanion: true})
	r, err = Build(idx, Deck{Source: "text", Entries: entries})
	if err != nil {
		t.Fatal(err)
	}
	if r.UnknownCopies != 2 || r.Companion != nil {
		t.Errorf("unknown copies %d, companion %+v; want 2 and none", r.UnknownCopies, r.Companion)
	}
}
