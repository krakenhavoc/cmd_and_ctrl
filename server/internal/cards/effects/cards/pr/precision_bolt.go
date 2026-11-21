package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Precision Bolt — Sorcery {2}{R} (Guilds of Ravnica, #5525):
//
//	"Precision Bolt deals 3 damage to any target."
//
// Lightning Bolt's body at sorcery speed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5f9ea6e4-f5be-4ea2-95b4-fbdc69685ac5",
		Name:         "Precision Bolt",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 3)), Answers: game.AnswerRemove},
		OnResolve:    effects.DamageToFirstTargetStep(3),
	})
}
