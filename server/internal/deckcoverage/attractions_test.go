package deckcoverage

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
)

// attractionEntries puts n distinct Attraction cards in idx and returns
// one deck row for each.
func attractionEntries(idx *cards.Index, n int) []deck.Entry {
	var out []deck.Entry
	for i := range n {
		name := fmt.Sprintf("Test Attraction %02d", i)
		idx.Put(cards.Card{ID: uuid.New(), OracleID: uuid.New(), Name: name,
			TypeLine: "Artifact — Attraction", Legalities: map[string]string{"commander": "legal"}})
		out = append(out, deck.Entry{Name: name, Count: 1})
	}
	return out
}

func violationCodes(r *Report) []string {
	var out []string
	for _, v := range r.Violations {
		out = append(out, v.Code)
	}
	return out
}

// ADR 0207 §2, CR 717.2: Attractions are the Attraction deck, outside
// the 100. They are reported on their own and the deck's size, singleton
// and counts are what they were without them.
func TestAttractionsAreReportedOutsideTheDeck(t *testing.T) {
	idx, tc := TestingIndex(t)
	base, err := Build(idx, Deck{Source: "text", Entries: testingDeck(tc)})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	entries := append(testingDeck(tc), attractionEntries(idx, 10)...)
	r, err := Build(idx, Deck{Source: "text", Entries: entries})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(r.Attractions) != 10 {
		t.Errorf("attractions = %d rows, want 10", len(r.Attractions))
	}
	if !reflect.DeepEqual(r.Cards, base.Cards) || !reflect.DeepEqual(r.Counts, base.Counts) || !reflect.DeepEqual(r.Copies, base.Copies) {
		t.Errorf("the Attractions changed the deck's own report")
	}
	if !reflect.DeepEqual(violationCodes(r), violationCodes(base)) {
		t.Errorf("violations with ten Attractions = %v, want the deck's own %v (no size or singleton complaint)",
			violationCodes(r), violationCodes(base))
	}
	if base.Attractions != nil {
		t.Errorf("a deck with none reports %+v", base.Attractions)
	}
}

// Fewer than ten, or a repeated name, is the warning seating gives
// (CR 717.2a), and only that one.
func TestShortOrRepeatedAttractionDeckWarnsLikeSeating(t *testing.T) {
	idx, tc := TestingIndex(t)
	base, _ := Build(idx, Deck{Source: "text", Entries: testingDeck(tc)})
	ten := attractionEntries(idx, 10)
	cases := map[string][]deck.Entry{
		"nine":                         ten[:9],
		"ten distinct plus one repeat": append(append([]deck.Entry{}, ten...), deck.Entry{Name: ten[0].Name, Count: 1}),
		"ten names, one twice over":    append(append([]deck.Entry{}, ten[1:]...), deck.Entry{Name: ten[0].Name, Count: 2}),
	}
	for name, extra := range cases {
		r, err := Build(idx, Deck{Source: "text", Entries: append(testingDeck(tc), extra...)})
		if err != nil {
			t.Fatalf("%s: Build: %v", name, err)
		}
		want := append(violationCodes(base), deck.CodeAttractionDeckIllegal)
		if !reflect.DeepEqual(violationCodes(r), want) {
			t.Errorf("%s: violations = %v, want %v", name, violationCodes(r), want)
		}
		if !reflect.DeepEqual(r.Cards, base.Cards) {
			t.Errorf("%s: the Attractions changed the deck's own cards", name)
		}
	}
}
