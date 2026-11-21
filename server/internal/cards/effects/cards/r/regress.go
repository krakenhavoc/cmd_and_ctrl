package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Regress — Instant {2}{U} (Mirrodin, #5606):
//
//	"Return target permanent to its owner's hand."
//
// Boomerang's text at a higher cost: the target is checked again as
// the spell resolves (CR 608.2b), and a token returned to a hand
// ceases to exist (CR 111.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2fa763ad-4d94-4c3a-a099-13ac22c09ce4",
		Name:         "Regress",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target permanent"),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve:    effects.BounceTheTarget,
	})
}
