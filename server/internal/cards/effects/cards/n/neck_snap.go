package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Neck Snap — Instant {3}{W} (Lorwyn, S82 OLD b557):
//
//	"Destroy target attacking or blocking creature."
//
// Puncturing Light without the power bound; the target is judged at announce
// and again on resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b8ad4276-759d-4d22-a47e-bd27edb34bff",
		Name:         "Neck Snap",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
		OnResolve:    effects.DestroyEachLegalTarget,
	})
}
