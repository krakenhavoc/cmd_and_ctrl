package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// aang_helpers.go — predicates and board queries for the Azorius
// flash / blink decklist. Kept out of helpers.go deliberately: that
// file is edited by every concurrent card batch, and none of this is
// needed outside these cards yet. Fold it in if a second deck wants
// the same queries.

// IsBasicLandWithSubtype returns a SearchLibrary predicate matching a
// basic land carrying `subtype` — "a basic Plains card" (Loyal
// Warhound), which neither existing helper can express:
// IsBasicLand admits any basic, and IsLandWithSubtype also admits
// nonbasics with the type (Hallowed Fountain is a Plains, but not a
// basic one).
//
// `subtype` MUST be lowercase. containsFoldASCII folds only the
// haystack byte, so a capitalised needle silently matches nothing.
// The sharper trap is IsBasicLandExcept(""), which looks like "any
// basic land" and matches NOTHING: an empty needle makes
// containsFoldASCII true for every card, so the negation always
// fails. Both failure modes are silent — a fetch that quietly finds
// no card rather than erroring.
//
// "Basic" is read off the SUPERTYPE (b30IsBasicLandCard), not
// IsBasicLand's "basic land" type-line substring: a Snow-Covered
// Plains is "Basic Snow Land — Plains", which is a basic Plains card
// and which the substring misses.
func IsBasicLandWithSubtype(subtype string) func(game.Card) bool {
	return func(c game.Card) bool {
		return b30IsBasicLandCard(c) && containsFoldASCII(c.TypeLine, subtype)
	}
}

// isActivePlayer reports whether playerID owns the current turn.
// "Whenever you cast a spell during an opponent's turn" (Brineborn
// Cutthroat) fires when this is false for the caster — the flash
// deck's core payoff shape.
func isActivePlayer(g *game.Game, playerID uuid.UUID) bool {
	if g == nil || g.Turn.ActiveSeat < 0 || g.Turn.ActiveSeat >= len(g.Seats) {
		return false
	}
	active := g.Seats[g.Turn.ActiveSeat]
	return active != nil && active.ID == playerID
}

// anOpponentHasMoreLands reports whether any player other than
// controller controls strictly more lands. Backs Loyal Warhound's
// intervening-if (CR 603.4), which is checked both when the trigger
// would go on the stack and again on resolution — so a land dropped
// in response correctly fizzles the fetch.
func anOpponentHasMoreLands(g *game.Game, controller uuid.UUID) bool {
	if g == nil {
		return false
	}
	counts := map[uuid.UUID]int{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsLand() {
			counts[c.Controller]++
		}
	}
	mine := counts[controller]
	for owner, n := range counts {
		if owner != controller && n > mine {
			return true
		}
	}
	return false
}
