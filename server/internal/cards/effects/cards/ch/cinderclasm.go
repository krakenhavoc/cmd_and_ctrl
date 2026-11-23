package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cinderclasm — Instant {1}{R} (Zendikar Rising, #5508):
//
//	"Kicker {R} (You may pay an additional {R} as you cast this spell.)
//	 Cinderclasm deals 1 damage to each creature. If it was kicked, it
//	 deals 2 damage to each creature instead."
//
// Fire Magic's sweep (damageEachMatching over every creature, whoever
// controls it), its size read off the cast (ctx.WasKicked). The
// "instead" replaces the 1, so a kicked cast deals 2, never 3.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "d7b49bee-2cb0-48e1-91c1-1e150a0da6ed",
		Name:          "Cinderclasm",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1}},
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{R}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			n := 1
			if ctx.WasKicked() {
				n = 2
			}
			return effects.DamageEachMatching(ctx, effects.Creature(), n)
		},
	})
}
