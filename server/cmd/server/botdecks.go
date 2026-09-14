package main

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/decks"
)

// botDeckCatalog is the wiring S31 sub-PR 4 promised and sub-PR 5
// never came back to connect: it adapts the curated archetype decks
// in internal/aiseat/decks to the aiseat.DeckSource the lobby's
// Add-bot picker reads.
//
// Until this existed, `main.go` passed aiseat.PlaceholderDecks() — a
// commander and ninety-nine Mountains, explicitly labelled "a stand-in
// until the curated archetype decks land (S31 sub-PR 5)". Those decks
// landed; the one-line swap did not, so every bot seated on a real
// deployment played mono-red basics while four tested, catalog-covered
// archetypes sat unreachable in the tree.
//
// It lives here rather than in internal/aiseat/decks for one reason:
// `decks` is the deck DATA and its coverage test, and it deliberately
// knows nothing about the lobby's wire shapes. The adapter is boot
// wiring, so it sits with the rest of the boot wiring.
type botDeckCatalog struct{}

// List returns every curated deck in picker order.
func (botDeckCatalog) List() []aiseat.DeckInfo {
	all := decks.All()
	out := make([]aiseat.DeckInfo, 0, len(all))
	for _, d := range all {
		out = append(out, botDeckInfo(d))
	}
	return out
}

// Decklist returns the plain-text decklist for id, in the dialect
// deck.ParseText accepts. The lobby then runs the same
// ParseText → Resolve → Validate pipeline a human's upload runs, so a
// curated deck that rots — a renamed card, a new ban — fails at
// exactly the place a player's would.
func (botDeckCatalog) Decklist(id string) (aiseat.DeckInfo, string, bool) {
	d, ok := decks.Lookup(id)
	if !ok {
		return aiseat.DeckInfo{}, "", false
	}
	return botDeckInfo(d), d.Decklist(), true
}

// botDeckInfo projects a curated deck onto the picker's wire shape.
//
// The archetype is prefixed onto the description rather than dropped:
// aiseat.DeckInfo has no Archetype field, and the deck NAMES are
// flavour ("Raid and Ransack", "Body Count"), so without it the picker
// is four names and no way to tell which one attacks. Issue #89 asked
// for an archetype picker; this is the archetype, carried in the one
// field that survives the interface.
func botDeckInfo(d decks.Deck) aiseat.DeckInfo {
	description := d.Summary
	if d.Archetype != "" {
		description = d.Archetype + " — " + d.Summary
	}
	colors := make([]string, 0, len(d.Identity))
	for _, c := range d.Identity {
		colors = append(colors, string(c))
	}
	return aiseat.DeckInfo{
		ID:          d.ID,
		Name:        d.Name,
		Description: description,
		Colors:      colors,
		Commander:   d.Commander.Name,
	}
}
