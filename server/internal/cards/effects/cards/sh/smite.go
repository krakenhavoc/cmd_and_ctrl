package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Smite — Instant {W} (Stronghold, #5627):
//
//	"Destroy target blocked creature."
//
// The shared BlockedCreature clause (CR 509.1h) and the
// destroyTheTargetPermanent body. Only an attacking creature that became
// blocked qualifies, so outside a combat with blocks there is no legal
// target and the spell can't be cast. A blocked creature stays blocked
// even if its blockers leave combat, so it is still a legal target at
// resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "367379f2-e1e4-48e7-b2c0-1003de55adbe",
		Name:         "Smite",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target blocked creature", effects.BlockedCreature()),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
