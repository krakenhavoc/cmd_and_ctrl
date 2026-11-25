package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Alley Evasion — Instant {W} (Aether Revolt, S82 older sets, #5536):
//
//	"Choose one —
//	 • Target creature you control gets +1/+2 until end of turn.
//	 • Return target creature you control to its owner's hand."
//
// Only the chosen bullet's target is chosen (CR 601.2b); a target that is
// gone as the spell resolves fizzles it (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b59c7c62-ce2f-4062-9790-566191aded69",
		Name:         "Alley Evasion",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.Mode("Target creature you control gets +1/+2 until end of turn.",
				effects.TargetCreature("target creature you control", effects.YouControl())),
				effects.ForTargets(effects.BenefitUntilEOT(0, 1, 2))),
			effects.ModeWithPurpose(effects.Mode("Return target creature you control to its owner's hand.",
				effects.TargetCreature("target creature you control", effects.YouControl())),
				game.Purpose{Answers: game.AnswerSurvive}),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.ModeTarget(ctx, 0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			switch {
			case ctx.HasMode(0):
				return effects.BoostUntilEOT{Target: t.ID, Power: 1, Toughness: 2, Label: "Alley Evasion — +1/+2 until end of turn"}.Apply(ctx)
			case ctx.HasMode(1):
				return effects.BounceToHand{Target: t.ID}.Apply(ctx)
			}
			return nil
		},
	})
}
