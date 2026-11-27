package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rescind — Instant {1}{U}{U} (Urza's Saga, S82 OLD b722):
//
//	Return target permanent to its owner's hand.
//	Cycling {2} ({2}, Discard this card: Draw a card.)
//
// Boomerang's bounce with cycling.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3f487c8e-a923-4031-b0a7-04f778544fe0",
		Name:         "Rescind",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target permanent"),
		OnResolve:    effects.BounceTheTarget,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
