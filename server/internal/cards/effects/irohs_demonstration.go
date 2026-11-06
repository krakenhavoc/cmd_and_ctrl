package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "8dd2e73a-d1b4-4d26-857f-9eb8e384c8ff",
		Name:         "Iroh's Demonstration",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(ModeDoing("Iroh's Demonstration deals 1 damage to each creature your opponents control.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return damageEachMatching(ctx, And(Creature(), OpponentControls()), 1)
				}),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, OpponentsOnly: true}}),
			ModeWithPurpose(ModeDoing("Iroh's Demonstration deals 4 damage to target creature.",
				TargetCreature("target creature"),
				DealFixedDamageToModesTarget(4)), ForTargets(DamageToTarget(0, 4))),
		),
	})
}
