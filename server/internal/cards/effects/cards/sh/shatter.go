package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shatter — Instant {1}{R} (Limited Edition Alpha, #5642):
//
//	"Destroy target artifact."
//
// One target clause and the shared "Destroy target [permanent]." body.
// Artifact() reads the post-layer type set, so an animated artifact is
// still a legal target and a creature that has lost the type is not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "84d32e68-72b9-432d-a023-7bc6a26b8ad0",
		Name:         "Shatter",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // #4734: destroys a permanent
		Targets:      effects.TargetPermanent("target artifact", effects.Artifact()),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
