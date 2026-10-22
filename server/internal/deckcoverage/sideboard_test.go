package deckcoverage

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
)

// ADR 0191, #3525: the sideboard is bucketed on its own, one row per
// distinct card, sorted like the deck's cards, and kept out of the deck's
// own rows and counts so a legal deck still totals 100.
func TestSideboardIsReportedOutsideTheDeck(t *testing.T) {
	idx, tc := TestingIndex(t)
	base, err := Build(idx, Deck{Source: "text", Entries: testingDeck(tc)})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if base.Sideboard != nil {
		t.Fatalf("a deck with no resolved sideboard reports one: %+v", base.Sideboard)
	}

	entries := append(testingDeck(tc),
		deck.Entry{Name: tc.Automated, Count: 2, IsSideboard: true},
		deck.Entry{Name: tc.Manual, Count: 1, IsSideboard: true},
		deck.Entry{Name: tc.Caveated, Count: 1, IsSideboard: true},
		deck.Entry{Name: tc.Unreviewed, Count: 1, IsSideboard: true},
		deck.Entry{Name: tc.Basic, Count: 1, IsSideboard: true},
		deck.Entry{Name: tc.Manual, Count: 1, IsSideboard: true}, // a second row of the same card
	)
	r, err := Build(idx, Deck{Source: "text", Entries: entries})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	type row struct {
		name   string
		bucket Bucket
		count  int
	}
	var got []row
	for _, c := range r.Sideboard {
		got = append(got, row{c.Name, c.Bucket, c.Count})
		if c.OracleID == "" {
			t.Errorf("%s has no oracle id", c.Name)
		}
		if c.Bucket == Caveats && (len(c.Caveats) == 0 || c.Caveats[0] != tc.CaveatText) {
			t.Errorf("caveats = %v, want the card's own %q", c.Caveats, tc.CaveatText)
		}
	}
	want := []row{
		{tc.Manual, Manual, 2},
		{tc.Unreviewed, Unreviewed, 1},
		{tc.Caveated, Caveats, 1},
		{tc.Automated, Automated, 2},
		{tc.Basic, NoEffect, 1},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("sideboard = %+v\nwant      %+v", got, want)
	}

	// The deck's own rows, counts and copies are what they were.
	if !reflect.DeepEqual(r.Cards, base.Cards) || !reflect.DeepEqual(r.Counts, base.Counts) ||
		!reflect.DeepEqual(r.Copies, base.Copies) || r.CopiesTotal() != base.CopiesTotal() {
		t.Errorf("the sideboard changed the deck's report:\n cards  %+v vs %+v\n counts %v vs %v\n copies %v vs %v",
			r.Cards, base.Cards, r.Counts, base.Counts, r.Copies, base.Copies)
	}
	// Listing it is not a reason to file a request.
	if r.Needed() != base.Needed() {
		t.Error("a sideboard card changed whether the deck needs a request")
	}
}

// An unresolvable sideboard name is in `unknown` and in no bucket, and
// does not count against the deck's size.
func TestSideboardUnknownNameIsNotBucketed(t *testing.T) {
	idx, tc := TestingIndex(t)
	entries := append(testingDeck(tc), deck.Entry{Name: "No Such Sideboard Card", Count: 3, IsSideboard: true})
	r, err := Build(idx, Deck{Source: "text", Entries: entries})
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range r.Sideboard {
		if strings.Contains(c.Name, "No Such") {
			t.Errorf("an unresolved card is in the sideboard: %+v", c)
		}
	}
	found := false
	for _, u := range r.Unknown {
		if u == "No Such Sideboard Card" {
			found = true
		}
	}
	if !found {
		t.Errorf("unknown = %v, want the sideboard name", r.Unknown)
	}
	if r.UnknownCopies != 2 {
		t.Errorf("unknown copies %d, want the deck's 2 and none of the sideboard's", r.UnknownCopies)
	}
}

// ADR 0191, decision 3 (2026-10-14): a sideboard card is not checked
// against the commander's colour identity or the banlist; that waits for
// game formats (#2352). A red card and a banned card in the sideboard of a
// green deck are listed and carry no violation, while the same cards in the
// 99 are violations (the control).
func TestSideboardIsNotCheckedAgainstColourIdentityOrTheBanlist(t *testing.T) {
	idx, tc := TestingIndex(t)
	legal := map[string]string{"commander": "legal"}
	idx.Put(cards.Card{ID: uuid.New(), OracleID: uuid.New(), Name: "Testing Red Spell", TypeLine: "Sorcery",
		ColorIdentity: []string{"R"}, Legalities: legal})
	idx.Put(cards.Card{ID: uuid.New(), OracleID: uuid.New(), Name: "Testing Banned Spell", TypeLine: "Sorcery",
		ColorIdentity: []string{"G"}, Legalities: map[string]string{"commander": "banned"}})

	hasViolation := func(r *Report, code, card string) bool {
		for _, v := range r.Violations {
			if v.Code == code && v.Card == card {
				return true
			}
		}
		return false
	}

	control, err := Build(idx, Deck{Source: "text", Entries: append(testingDeck(tc),
		deck.Entry{Name: "Testing Red Spell", Count: 1},
		deck.Entry{Name: "Testing Banned Spell", Count: 1})})
	if err != nil {
		t.Fatal(err)
	}
	if !hasViolation(control, deck.CodeColorIdentity, "Testing Red Spell") ||
		!hasViolation(control, deck.CodeNotLegalInFormat, "Testing Banned Spell") {
		t.Fatalf("the control (both cards in the deck) raised no violation: %+v", control.Violations)
	}

	r, err := Build(idx, Deck{Source: "text", Entries: append(testingDeck(tc),
		deck.Entry{Name: "Testing Red Spell", Count: 1, IsSideboard: true},
		deck.Entry{Name: "Testing Banned Spell", Count: 1, IsSideboard: true})})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Sideboard) != 2 {
		t.Fatalf("sideboard = %+v, want both cards listed", r.Sideboard)
	}
	for _, v := range r.Violations {
		if v.Card == "Testing Red Spell" || v.Card == "Testing Banned Spell" {
			t.Errorf("a sideboard card was checked: %+v", v)
		}
	}
}

// The JSON field is additive: absent for a deck with no sideboard.
func TestSideboardJSONIsOmittedWhenEmpty(t *testing.T) {
	idx, tc := TestingIndex(t)
	r, err := Build(idx, Deck{Source: "text", Entries: testingDeck(tc)})
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(r)
	if strings.Contains(string(raw), `"sideboard"`) {
		t.Errorf("an empty sideboard is in the JSON: %s", raw)
	}
	entries := append(testingDeck(tc), deck.Entry{Name: tc.Vanilla, Count: 1, IsSideboard: true})
	r, err = Build(idx, Deck{Source: "text", Entries: entries})
	if err != nil {
		t.Fatal(err)
	}
	raw, _ = json.Marshal(r)
	if !strings.Contains(string(raw), `"sideboard":[{"name":"`+tc.Vanilla+`"`) {
		t.Errorf("the sideboard is not in the JSON: %s", raw)
	}
}
