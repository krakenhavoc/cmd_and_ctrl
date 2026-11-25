package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Merciless Resolve — Instant {2}{B} (Shadows over Innistrad, #5543):
//
//	"As an additional cost to cast this spell, sacrifice a creature or
//	 land.
//	 Draw two cards."
//
// The sacrifice is an additional cost paid as the spell is cast
// (CR 601.2b, CR 601.2h): a pool with no creature or land to give refuses
// the cast.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "e9eecc10-8821-4b61-a042-01fef435d37c",
		Name:           "Merciless Resolve",
		Completeness:   effects.CompletenessFull,
		Purpose:        game.Purpose{Draws: 2, Answers: game.AnswerSacOutlet},
		AdditionalCost: effects.SacrificeCost("a creature or land", effects.Or(effects.Creature(), effects.Land())),
		OnResolve:      effects.ResolveSteps(effects.DrawCards{N: 2}),
	})
}
