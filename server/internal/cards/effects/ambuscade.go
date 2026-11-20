package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ambuscade — Instant {2}{G} (Hour of Devastation, #5534):
//
//	"Target creature you control gets +1/+0 until end of turn. It deals
//	 damage equal to its power to target creature an opponent controls."
//
// Huatli's Final Strike's text: two clauses, the boost first, then the
// shared one-sided bite, which reads the boosted power. Each target is
// judged on its own at resolution (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5e419975-e054-4c35-8845-a08589ebb08f",
		Name:         "Ambuscade",
		Completeness: CompletenessFull,
		Purpose: game.Purpose{Answers: game.AnswerRemove, // a one-sided fight, as Bite Down (#4734)
			Targets: game.ForTargets(BenefitUntilEOT(0, 1, 0))},
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature an opponent controls", OpponentControls()),
		),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				mine, ok := ctx.ClauseTarget(0)
				if !ok || mine.Kind != game.TargetCard {
					return nil
				}
				return BoostUntilEOT{Target: mine.ID, Power: 1, Label: "Ambuscade — +1/+0 until end of turn"}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error { return oneSidedBite(ctx.Item, ctx) }),
		),
	})
}
