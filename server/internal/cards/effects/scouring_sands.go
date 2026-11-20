package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "f7f32fec-b384-4021-a69f-2dc9220ebc64",
		Name:         "Scouring Sands",
		Completeness: CompletenessFull,
		Purpose: game.Purpose{
			Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, OpponentsOnly: true},
			Scry:  1,
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := damageEachMatching(ctx, And(Creature(), OpponentControls()), 1); err != nil {
				return err
			}
			return Scry{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
