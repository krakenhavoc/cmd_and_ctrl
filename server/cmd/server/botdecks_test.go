package main

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/decks"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
)

// The adapter is only useful if it is the interface the lobby holds.
var _ aiseat.DeckSource = botDeckCatalog{}

// TestBotDeckCatalogListsEveryCuratedDeck is the regression that
// matters: a deck added to internal/aiseat/decks and NOT reachable
// from the picker is the exact failure this file was written to fix,
// and it is invisible without an assertion.
func TestBotDeckCatalogListsEveryCuratedDeck(t *testing.T) {
	got := (botDeckCatalog{}).List()
	want := decks.IDs()
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
			total += e.Count
			if e.IsCommander {
				commanders += e.Count
			}
		}
		if total != 100 {
			t.Errorf("%s: decklist parses to %d cards, want 100", info.ID, total)
		}
		if commanders != 1 {
			t.Errorf("%s: %d commanders, want 1", info.ID, commanders)
		}
	}
}

func TestBotDeckCatalogUnknownID(t *testing.T) {
	if _, _, ok := (botDeckCatalog{}).Decklist("no-such-deck"); ok {
		t.Error("an unknown deck id must not resolve — the lobby turns the false into a 422")
	}
}
