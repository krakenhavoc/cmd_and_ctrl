package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Terror — Instant {1}{B} (Limited Edition Alpha, #5642):
//
//	"Destroy target nonartifact, nonblack creature. It can't be
//	 regenerated."
//
// Doom Blade's colour clause and Go for the Throat's type clause on one
// target, with Terminate's body: "It can't be regenerated" rides the
// destroy onto the CR 614 event and a regeneration shield declines it
// (CR 701.19c, #667). Both clauses read the post-layer characteristics,
// so a creature that turns black or becomes an artifact in response is
// no longer a legal target at resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b81f041d-98db-4408-9472-c483e4a502bc",
		Name:         "Terror",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // #4734: destroys a permanent
		Targets:      effects.TargetCreature("target nonartifact, nonblack creature", effects.Not(effects.Artifact()), effects.NonBlack()),
		OnResolve:    effects.DestroyTheTargetPermanentNoRegen,
	})
}
