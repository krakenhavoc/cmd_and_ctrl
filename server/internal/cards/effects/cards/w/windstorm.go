package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Windstorm — Instant {X}{G} (Khans of Tarkir):
//
//	"Windstorm deals X damage to each creature with flying."
//
// Squall's each-creature-with-flying sweep at X: one batch of damage from the
// spell to each creature that has flying through the layers as it resolves,
// yours included. X = 0 deals nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9969b822-2ead-4956-bac2-875ffd90e872",
		Name:         "Windstorm",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		// ADR 0126 §6: only creatures with flying.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := ctx.X()
			if x <= 0 {
				return nil
			}
			return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.HasKeyword("flying")), x)
		},
	})
}
