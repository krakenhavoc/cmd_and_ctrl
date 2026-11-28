package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lava Storm — {3}{R}{R} Instant (Weatherlight, S82 OLD batch 771, #5629):
//
//	"Lava Storm deals 2 damage to each attacking creature or Lava Storm deals 2
//	 damage to each blocking creature."
//
// A modal spell with one of two untargeted modes, chosen as it is cast (CR
// 700.2): 2 damage to each attacking creature (Hail Storm's and Sandstorm's
// attackers), or to each blocking creature. The creatures are those attacking or
// blocking as the spell resolves, any player's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "78d33345-a147-402f-9a40-1e8260d7d44e",
		Name:         "Lava Storm",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{
			Answers: game.AnswerRemove,
			Sweep:   game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2, Partial: true},
		},
		Modes: effects.ChooseOne(
			effects.ModeDoing("Lava Storm deals 2 damage to each attacking creature.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.DamageEachMatching(ctx, effects.AttackingCreature(), 2)
				}),
			effects.ModeDoing("Lava Storm deals 2 damage to each blocking creature.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.DamageEachMatching(ctx, effects.BlockingCreature(), 2)
				}),
		),
	})
}
