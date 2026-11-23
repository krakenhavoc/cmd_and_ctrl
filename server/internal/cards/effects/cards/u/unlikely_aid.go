package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unlikely Aid — Instant {1}{B} (War of the Spark, #5520):
//
//	"Target creature gets +2/+0 and gains indestructible until end of turn. (Damage and effects that say "destroy" don't destroy it.)"
//
// Adamant Will's body with +2/+0.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "795e0ffb-fcd9-4cb9-ad18-8a0217229bfe",
		Name:         "Unlikely Aid",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{Answers: game.AnswerSurvive | game.AnswerPump,
			Targets: game.ForTargets(effects.BenefitUntilEOT(0, 2, 0, "indestructible"))},
		Targets: effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.PumpAndGrantFirstTarget(2, 0, "Unlikely Aid — +2/+0 and indestructible", "indestructible")(ctx.Game, item)
		},
	})
}
