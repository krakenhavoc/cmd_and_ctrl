package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "5e419975-e054-4c35-8845-a08589ebb08f",
		Name:         "Ambuscade",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{Answers: game.AnswerRemove, // a one-sided fight, as Bite Down (#4734)
			Targets: game.ForTargets(effects.BenefitUntilEOT(0, 1, 0))},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				mine, ok := ctx.ClauseTarget(0)
				if !ok || mine.Kind != game.TargetCard {
					return nil
				}
				return effects.BoostUntilEOT{Target: mine.ID, Power: 1, Label: "Ambuscade — +1/+0 until end of turn"}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error { return effects.OneSidedBite(ctx.Item, ctx) }),
		),
	})
}
