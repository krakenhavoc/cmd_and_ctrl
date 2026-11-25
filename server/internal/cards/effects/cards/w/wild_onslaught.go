package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wild Onslaught — Instant {3}{G} (Dominaria):
//
//	"Kicker {4} (You may pay an additional {4} as you cast this spell.)
//	 Put a +1/+1 counter on each creature you control. If this spell was kicked,
//	 put two +1/+1 counters on each creature you control instead."
//
// One placement per creature of one or two counters (CR 122.1, 614): the
// creatures are those you control as the spell resolves, each getting its
// counters together, so a counter-doubling effect applies once to each.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "7ca9ef3e-e5b2-4eb5-87b3-0d97637428a0",
		Name:          "Wild Onslaught",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Answers: game.AnswerPump},
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{4}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			n := 1
			if ctx.WasKicked() {
				n = 2
			}
			return effects.B11PutCountersOnEachCreatureYouControl(ctx.Game, item, n)
		},
	})
}
