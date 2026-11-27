package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lightning Blast — Instant {3}{R} (Eighth Edition, S82 OLD batch 610, #5607):
//
//	"Lightning Blast deals 4 damage to any target."
//
// Lightning Bolt's shape at four.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "91fd731c-e076-4f2d-9f22-872880c3cc3d",
		Name:         "Lightning Blast",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 4)), Answers: game.AnswerRemove},
		OnResolve:    effects.DamageToFirstTargetStep(4),
	})
}
