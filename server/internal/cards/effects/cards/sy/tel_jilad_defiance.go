package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tel-Jilad Defiance — Instant {1}{G} (Scars of Mirrodin, #5577):
//
//	"Target creature gains protection from artifacts until end of turn.
//	 Draw a card."
//
// Expedite's grant-then-draw with protection from artifacts
// (ebTargetCreatureGainsUntilEOT); protection is the engine's
// (CR 702.16).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "73c7d938-5cec-4f96-b649-3f67bef29407",
		Name:         "Tel-Jilad Defiance",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{Answers: game.AnswerProtect, Draws: 1,
			Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "protection from artifacts"))},
		Targets: effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.EbTargetCreatureGainsUntilEOT("protection from artifacts", "Tel-Jilad Defiance — protection from artifacts")(ctx.Game, ctx.Item)
			}),
			effects.DrawCards{N: 1},
		),
	})
}
