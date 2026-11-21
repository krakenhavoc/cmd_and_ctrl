package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scouring Sands — Sorcery {1}{R} (Born of the Gods,
// #5559):
//
//	"Scouring Sands deals 1 damage to each creature your opponents
//	 control. Scry 1. (Look at the top card of your library. You may put
//	 that card on the bottom.)"
//
// Smash to Dust's third mode (damageEachMatching over the opponents'
// creatures), then the scry.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f7f32fec-b384-4021-a69f-2dc9220ebc64",
		Name:         "Scouring Sands",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{
			Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, OpponentsOnly: true},
			Scry:  1,
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.OpponentControls()), 1); err != nil {
				return err
			}
			return effects.Scry{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
