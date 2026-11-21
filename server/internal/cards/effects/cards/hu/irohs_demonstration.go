package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Iroh's Demonstration — Sorcery — Lesson {1}{R}:
//
//	"Choose one —
//	 • Iroh's Demonstration deals 1 damage to each creature your
//	   opponents control.
//	 • Iroh's Demonstration deals 4 damage to target creature."
//
// The first bullet's set is read as it resolves; the second fizzles on
// an illegal target (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8dd2e73a-d1b4-4d26-857f-9eb8e384c8ff",
		Name:         "Iroh's Demonstration",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Iroh's Demonstration deals 1 damage to each creature your opponents control.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.OpponentControls()), 1)
				}),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, OpponentsOnly: true}}),
			effects.ModeWithPurpose(effects.ModeDoing("Iroh's Demonstration deals 4 damage to target creature.",
				effects.TargetCreature("target creature"),
				effects.DealFixedDamageToModesTarget(4)), effects.ForTargets(effects.DamageToTarget(0, 4))),
		),
	})
}
