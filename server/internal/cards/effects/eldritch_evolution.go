package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eldritch Evolution — Sorcery {1}{G}{G}:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Search your library for a creature card with mana value X or less,
//	 where X is 2 plus the sacrificed creature's mana value. Put that
//	 card onto the battlefield, then shuffle. Exile Eldritch Evolution."
//
// X is 2 plus the sacrificed creature's mana value as it last existed on
// the battlefield (CR 608.2h), read off the payment record (ADR 0113
// §1): a token that is no copy is 0, so X is 2 (the 2016-07-13 rulings).
// A card in the library with {X} in its cost has X = 0.
//
// "Exile Eldritch Evolution" is done as the spell finishes resolving,
// so it goes to exile rather than the graveyard whether or not a
// creature card was found.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:       "0f77c0c9-4dc4-489a-b547-e93287c4d1a5",
		Name:           "Eldritch Evolution",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("a creature", Creature()),
		OnResolve: ResolveSteps(
			// The exile of this spell waits for the search (CR 608.2c, #4155).
			Instruction(func(ctx *Context) error {
				x := 2 + sacrificedManaValue(ctx)
				return SearchLibrary{
					Player: ctx.Controller(),
					Predicate: func(c game.Card) bool {
						return c.IsCreature() && c.ManaValue() <= x
					},
					Dest:    game.ZoneBattlefield,
					Limit:   1,
					Shuffle: true,
					Reason:  "Eldritch Evolution — a creature card with mana value " + strconv.Itoa(x) + " or less",
				}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error { return ExileTarget{Target: ctx.Source()}.Apply(ctx) }),
		),
	})
}
