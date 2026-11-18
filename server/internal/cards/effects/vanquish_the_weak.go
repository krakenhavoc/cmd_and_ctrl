package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vanquish the Weak — Instant {2}{B}:
//
//	"Destroy target creature with power 3 or less."
//
// The power bound is a target restriction (PowerLE), judged at announce
// and again as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4a3cc4f6-037f-461d-a3b7-9f9e24c4b8cf",
		Name:         "Vanquish the Weak",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature with power 3 or less", PowerLE(3)),
		OnResolve:    destroyEachLegalTarget,
	})
}
