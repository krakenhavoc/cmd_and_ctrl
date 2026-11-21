package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flame Lash — Instant {3}{R} (Bloomburrow, #4467):
//
//	"Flame Lash deals 4 damage to any target."
//
// Lightning Strike's any-target clause with four damage, the spell as
// its source. A target gone by resolution takes nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c65e804b-9222-4fa8-88d1-df06a3279f22",
		Name:         "Flame Lash",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 4)), Answers: game.AnswerRemove},
		OnResolve:    effects.DamageToFirstTargetStep(4),
	})
}
