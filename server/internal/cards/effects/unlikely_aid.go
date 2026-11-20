package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unlikely Aid — Instant {1}{B} (War of the Spark, #5520):
//
//	"Target creature gets +2/+0 and gains indestructible until end of turn. (Damage and effects that say "destroy" don't destroy it.)"
//
// Adamant Will's body with +2/+0.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "795e0ffb-fcd9-4cb9-ad18-8a0217229bfe",
		Name:         "Unlikely Aid",
		Completeness: CompletenessFull,
		Purpose: game.Purpose{Answers: game.AnswerSurvive | game.AnswerPump,
			Targets: game.ForTargets(BenefitUntilEOT(0, 2, 0, "indestructible"))},
		Targets: TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return pumpAndGrantFirstTarget(2, 0, "Unlikely Aid — +2/+0 and indestructible", "indestructible")(ctx.Game, item)
		},
	})
}
