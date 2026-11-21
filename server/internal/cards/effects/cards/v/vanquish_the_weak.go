package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vanquish the Weak — Instant {2}{B}:
//
//	"Destroy target creature with power 3 or less."
//
// The power bound is a target restriction (PowerLE), judged at announce
// and again as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4a3cc4f6-037f-461d-a3b7-9f9e24c4b8cf",
		Name:         "Vanquish the Weak",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature with power 3 or less", effects.PowerLE(3)),
		OnResolve:    effects.DestroyEachLegalTarget,
	})
}
