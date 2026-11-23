package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wrecking Ball — Instant {2}{B}{R} (Dissension, #5598):
//
//	"Destroy target creature or land."
//
// Fissure's target without its "can't be regenerated".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0a1b0910-3b0c-4aa3-b47f-8eb069122743",
		Name:         "Wrecking Ball",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target creature or land", effects.Or(effects.Creature(), effects.Land())),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
