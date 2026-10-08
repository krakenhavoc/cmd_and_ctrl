package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Neoform — Sorcery {G}{U}:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Search your library for a creature card with mana value equal to 1
//	 plus the sacrificed creature's mana value, put that card onto the
//	 battlefield with an additional +1/+1 counter on it, then shuffle."
//
// The mana value is the sacrificed creature's as it last existed on the
// battlefield (CR 608.2h), read off the payment record (ADR 0113 §1); a
// token that is no copy is 0, so the search is for mana value 1. A card
// in the library with {X} in its cost has X = 0.
//
// The counter is seeded on the entry event (ADR 0113, third amendment of 2026-10-08, #2098), so it is
// there as the creature enters: an enters trigger that reads its counters
// sees it, and Doubling Season doubles it. "Additional" means a creature
// that enters with counters of its own keeps them.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:       "420c6dcf-966d-4a4c-a0ef-23037ab8b325",
		Name:           "Neoform",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("a creature", Creature()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			mv := 1 + sacrificedManaValue(ctx)
			return SearchLibrary{
				Player: item.Controller,
				Predicate: func(c game.Card) bool {
					return c.IsCreature() && c.ManaValue() == mv
				},
				Dest:               game.ZoneBattlefield,
				Limit:              1,
				Shuffle:            true,
				EntersWithCounters: map[string]int{"+1/+1": 1},
				Reason:             "Neoform — a creature card with mana value " + strconv.Itoa(mv),
			}.Apply(ctx)
		},
	})
}
