package main

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/decks"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The adapter is only useful if it is the interface the lobby holds.
var _ aiseat.DeckSource = botDeckCatalog{}

// TestBotDeckCatalogListsEveryCuratedDeck is the regression that
// matters: a deck added to internal/decks and NOT reachable
// from the picker is the exact failure this file was written to fix,
// and it is invisible without an assertion.
func TestBotDeckCatalogListsEveryCuratedDeck(t *testing.T) {
	got := (botDeckCatalog{}).List()
	// Every format's decks (ADR 0222 slice 1): the Commander decks in
	// their order, then the others, each row naming its format.
	var want []string
	for _, d := range decks.Every() {
		want = append(want, d.ID)
		for _, info := range got {
			if info.ID != d.ID {
				continue
			}
			wantFormat := ""
			if d.FormatOf() != game.FormatCommander {
				wantFormat = string(d.FormatOf())
			}
			if info.Format != wantFormat {
				t.Errorf("%s: format %q on the wire, want %q", d.ID, info.Format, wantFormat)
			}
		}
	}
	if n := len(decks.IDs()); len(got) < n || got[n-1].ID != decks.IDs()[n-1] {
		t.Errorf("the Commander decks are not the picker's first %d rows, in their order", n)
	}
	if len(got) != len(want) {
		t.Fatalf("picker offers %d decks, catalog has %d", len(got), len(want))
	}
	if len(want) == 0 {
		t.Fatal("the curated catalog is empty — the picker would have nothing to offer")
	}
	for i, info := range got {
		if info.ID != want[i] {
			t.Errorf("deck %d: picker order %q, catalog order %q", i, info.ID, want[i])
		}
	}
}

// TestBotDeckCatalogIsNotThePlaceholder pins the swap itself. The
// placeholder source is still in the tree (it is a legitimate fuzzing
// deck), so nothing stops a future edit from wiring it back.
func TestBotDeckCatalogIsNotThePlaceholder(t *testing.T) {
	placeholder := map[string]bool{}
	for _, info := range aiseat.PlaceholderDecks().List() {
		placeholder[info.ID] = true
	}
	for _, info := range (botDeckCatalog{}).List() {
		if placeholder[info.ID] {
			t.Errorf("the picker is still offering the placeholder deck %q", info.ID)
		}
	}
}

// TestBotDeckCatalogInfoIsRenderable checks the fields the Add-bot
// picker actually draws. A deck whose archetype never reaches the
// wire leaves the player choosing between four flavour names.
func TestBotDeckCatalogInfoIsRenderable(t *testing.T) {
	for _, info := range (botDeckCatalog{}).List() {
		d, ok := decks.Lookup(info.ID)
		if !ok {
			t.Fatalf("%s: listed but not in the catalog", info.ID)
		}
		if info.Name == "" {
			t.Errorf("%s: no display name", info.ID)
		}
		if d.Archetype != "" && !strings.HasPrefix(info.Description, d.Archetype) {
			t.Errorf("%s: description %q does not lead with archetype %q",
				info.ID, info.Description, d.Archetype)
		}
		if info.Commander != d.Commander.Name {
			t.Errorf("%s: commander %q, want %q", info.ID, info.Commander, d.Commander.Name)
		}
		if len(info.Colors) != len([]rune(d.Identity)) {
			t.Errorf("%s: %d colour pips for identity %q", info.ID, len(info.Colors), d.Identity)
		}
		for _, c := range info.Colors {
			if len(c) != 1 || !strings.Contains("WUBRG", c) {
				t.Errorf("%s: %q is not a WUBRG letter", info.ID, c)
			}
		}
	}
}

// TestBotDeckCatalogDecklistParses runs the decklist bytes the lobby
// will hand to resolveDeckSource through the same parser a player's
// upload hits. It stops short of Resolve/Validate, which need a
// Scryfall index — decks_test.go's CMDCTRL_SCRYFALL_DUMP-gated test
// owns that half.
func TestBotDeckCatalogDecklistParses(t *testing.T) {
	src := botDeckCatalog{}
	for _, info := range src.List() {
		gotInfo, list, ok := src.Decklist(info.ID)
		if !ok {
			t.Fatalf("%s: listed but Decklist says unknown", info.ID)
		}
		if gotInfo.ID != info.ID {
			t.Errorf("%s: Decklist returned info for %q", info.ID, gotInfo.ID)
		}
		entries, err := deck.ParseText(list)
		if err != nil {
			t.Fatalf("%s: ParseText: %v", info.ID, err)
		}
		total, commanders := 0, 0
		for _, e := range entries {
			if e.IsCompanion {
				continue // outside the game, not one of the 100 (ADR 0172)
			}
			total += e.Count
			if e.IsCommander {
				commanders += e.Count
			}
		}
		// A Commander deck is a commander and 99; a deck of a format
		// with no commander is the format's minimum, exactly.
		spec := game.Format(info.Format).Spec()
		wantTotal, wantCommanders := spec.Deck.MinCards, 0
		if spec.CommanderExpected {
			wantCommanders = 1
		}
		if total != wantTotal {
			t.Errorf("%s: decklist parses to %d cards, want %d", info.ID, total, wantTotal)
		}
		if commanders != wantCommanders {
			t.Errorf("%s: %d commanders, want %d", info.ID, commanders, wantCommanders)
		}
	}
}

func TestBotDeckCatalogUnknownID(t *testing.T) {
	if _, _, ok := (botDeckCatalog{}).Decklist("no-such-deck"); ok {
		t.Error("an unknown deck id must not resolve — the lobby turns the false into a 422")
	}
}
