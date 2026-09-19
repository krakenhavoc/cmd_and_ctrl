package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// wire.go connects the catalog to the engine: one assignment, at
// package init. Importing this package (the blank import from
// main.go is enough) switches the game's resolution path from
// manual-sandbox-only to catalog-aware.
//
// Until #622 this file set twenty-two per-slot hooks, each doing its
// own registry lookup and several re-projecting slices per call. The
// engine now reads one precomputed *game.CardDef per card, built by
// buildDef at Register; the per-slot variables in the game package
// default to reading it and remain only as test seams. Absence of a
// catalog entry still matters (lookupDef returns nil) and is what
// preserves the opt-in invariant for non-catalog cards.
//
// Added in S14 sub-PR 3; collapsed in #622.

func init() {
	game.CatalogLookup = lookupDef
}

// selfOnly is the AppliesTo predicate for the synthesized
// printed-keyword StaticAbility — the keyword list applies only to
// the card itself, not to any other battlefield card. Separated
// here so the closure inside the hook is a bare reference rather
// than an inline func literal per call.
func selfOnly(target *game.Card, g *game.Game, source *game.Card) bool {
	return target.InstanceID == source.InstanceID
}

// keywordSliceContains is a small helper for the synthesized
// keyword apply — avoids re-appending a keyword already in the
// slice (e.g. Lord of Atlantis grants "islandwalk" to a Merfolk
// that already has printed islandwalk from its own
// `PrintedKeywords`). Named to avoid colliding with the
// `containsString` helper in test files in this package.
func keywordSliceContains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}

// appendKeywordsTo is the layer-6 Apply body every self-only keyword
// grant shares: append each keyword the object does not already have.
//
// Two callers, which is why it is named — the printed-keyword static
// buildDef synthesises for every card with a PrintedKeywords list, and
// ThresholdKeywords, a station card's "{N+} | Flying" line. The
// membership check is the load-bearing half: a permanent that already
// has the keyword must not end up with it twice, because the wire
// renders the badge list verbatim.
func appendKeywordsTo(c *game.Characteristic, keywords []string) {
	for _, kw := range keywords {
		if !keywordSliceContains(c.Abilities, kw) {
			c.Abilities = append(c.Abilities, kw)
		}
	}
}
