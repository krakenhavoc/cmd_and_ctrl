package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vanquish — Instant {2}{W} (Fifth Dawn, #5604):
//
//	"Destroy target blocking creature."
//
// Ride Down's "target blocking creature" clause on Shatter's destroy.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "683ec6d2-4ea7-46a7-911e-12c83e4c4a06",
		Name:         "Vanquish",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target blocking creature", effects.BlockingCreature()),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
