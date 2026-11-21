package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Immolating Glare — Instant {1}{W} (Oath of the Gatewatch, #5545):
//
//	"Destroy target attacking creature."
//
// Kill Shot's target clause (an attacking creature, judged at announce
// and at resolution) and Bedevil's body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "046b6989-5b56-4b52-a655-109b1eb26a93",
		Name:         "Immolating Glare",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
