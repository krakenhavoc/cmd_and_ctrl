package decks

import (
	"sort"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// standard_test.go holds the curated Standard decks (standard.go) to
// everything that can be checked without a Scryfall dump. The card
// pool (legalities.standard) needs the dump:
// realdump_standard_manual_test.go.

// TestThereAreStandardDecksForABot is the slice's promise (ADR 0222
// slice 1): at least two curated Standard decks, so a person can play
// Standard against a bot.
func TestThereAreStandardDecksForABot(t *testing.T) {
	if got := len(In(game.FormatStandard)); got < 2 {
		t.Fatalf("%d curated Standard decks, want at least 2", got)
	}
	for _, d := range In(game.FormatStandard) {
		if d.HasCommander() {
			t.Errorf("%s: a Standard deck has no commander, but names %q", d.ID, d.Commander.Name)
		}
	}
}

// TestCommanderCatalogIsUnchangedByFormats pins that the Commander
// catalog means what it meant before formats: All and IDs are the
// Commander decks and nothing else, and each still has its commander.
func TestCommanderCatalogIsUnchangedByFormats(t *testing.T) {
	ids := map[string]bool{}
	for _, d := range All() {
		ids[d.ID] = true
		if d.FormatOf() != game.FormatCommander {
			t.Errorf("All() lists %s, a %s deck", d.ID, d.FormatOf())
		}
		if !d.HasCommander() {
			t.Errorf("%s has no commander", d.ID)
		}
	}
	for _, id := range IDs() {
		if !ids[id] {
			t.Errorf("IDs() lists %s, which All() does not", id)
		}
	}
	if got, want := len(In(game.FormatCommander)), len(All()); got != want {
		t.Errorf("In(commander) has %d decks, All() has %d", got, want)
	}
	if got, want := len(In("")), len(All()); got != want {
		t.Errorf("In(\"\") has %d decks, want the %d Commander decks: an empty format is Commander", got, want)
	}
}

// TestEveryCatalogIDIsUnique: a deck is seated by ID across formats.
func TestEveryCatalogIDIsUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, d := range Every() {
		if seen[d.ID] {
			t.Errorf("deck ID %q is used twice", d.ID)
		}
		seen[d.ID] = true
		got, ok := Lookup(d.ID)
		if !ok || got.ID != d.ID {
			t.Errorf("Lookup(%q) = %v, %v", d.ID, got.ID, ok)
		}
	}
}

// TestStandardDecksAreLegalByTheNumbers checks CR 100.2a offline: "each
// deck has a minimum deck size of 60 cards" and "no more than four of
// any card with a particular English name other than basic land cards".
// The curated decks are exactly sixty.
func TestStandardDecksAreLegalByTheNumbers(t *testing.T) {
	rules := game.FormatStandard.Spec().Deck
	for _, d := range In(game.FormatStandard) {
		t.Run(d.ID, func(t *testing.T) {
			if got := d.Size(); got != rules.MinCards {
				t.Errorf("deck has %d cards, want %d", got, rules.MinCards)
			}
			copies := map[string]int{}
			lands := 0
			for _, c := range d.Mainboard {
				if !c.Basic {
					copies[c.Name] += c.Copies()
					continue
				}
				lands += c.Copies()
				if !basicLandNames[c.Name] {
					t.Errorf("%s is marked Basic but is not a basic land", c.Name)
				}
			}
			for name, n := range copies {
				if n > rules.MaxCopies {
					t.Errorf("%d copies of %s, at most %d allowed", n, name, rules.MaxCopies)
				}
			}
			if lands < 20 || lands > 26 {
				t.Errorf("%d lands in a sixty-card deck; want 20 to 26", lands)
			}
			for _, c := range d.Mainboard {
				for _, sym := range c.Identity {
					if !strings.ContainsRune(d.Identity, sym) {
						t.Errorf("%s has identity %s, outside the deck's %s: it could not be cast off the deck's basics", c.Name, c.Identity, d.Identity)
					}
				}
			}
		})
	}
}

// TestStandardCardsAreRegisteredAndFull is the bar the Commander decks
// are held to (TestEveryCardResolvesToARegisteredSpec,
// TestEveryCuratedCardIsFull), with no minor-caveats list: every
// non-basic card is in the catalog under the declared oracle ID and
// name, and graded full.
func TestStandardCardsAreRegisteredAndFull(t *testing.T) {
	for _, d := range In(game.FormatStandard) {
		t.Run(d.ID, func(t *testing.T) {
			var bad []string
			for _, c := range d.Mainboard {
				if c.Basic {
					continue
				}
				spec, ok := effects.Lookup(c.OracleID)
				switch {
				case !ok:
					bad = append(bad, c.Name+": no registered effects.Spec for "+c.OracleID)
				case spec.Name != c.Name:
					bad = append(bad, c.Name+": the catalog calls "+c.OracleID+" "+spec.Name)
				case spec.Completeness != effects.CompletenessFull:
					bad = append(bad, c.Name+": graded "+spec.Completeness.String()+", not full")
				}
			}
			if len(bad) > 0 {
				sort.Strings(bad)
				t.Fatalf("%d card(s) fail the curated bar:\n\t%s", len(bad), strings.Join(bad, "\n\t"))
			}
			if cov := CoverageOf(nil, d); !cov.Complete() {
				t.Errorf("coverage is not complete: %+v", cov)
			}
		})
	}
}

// TestStandardDecklistParsesWithNoCommander: the text a Standard deck
// emits is sixty mainboard cards, no commander and no sideboard, through
// the parse a player's paste takes.
func TestStandardDecklistParsesWithNoCommander(t *testing.T) {
	for _, d := range In(game.FormatStandard) {
		t.Run(d.ID, func(t *testing.T) {
			entries, err := deck.ParseText(d.Decklist())
			if err != nil {
				t.Fatalf("ParseText: %v", err)
			}
			entries = deck.EntriesFor(d.FormatOf(), entries)
			mainboard := 0
			for _, e := range entries {
				if e.IsCommander || e.IsSideboard || e.IsCompanion {
					t.Errorf("%s is not a mainboard row: %+v", e.Name, e)
					continue
				}
				mainboard += e.Count
			}
			if mainboard != 60 {
				t.Errorf("parsed %d mainboard cards, want 60", mainboard)
			}
		})
	}
}
