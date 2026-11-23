package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Searing Spear — Instant {1}{R} (Magic 2013, #5568):
//
//	"Searing Spear deals 3 damage to any target."
//
// Lightning Strike at instant speed: the same any-target clause and the
// same three damage from the spell as its source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aafd44c1-74a9-4aa7-a4a2-67eb39a07478",
		Name:         "Searing Spear",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 3)), Answers: game.AnswerRemove},
		OnResolve:    effects.DamageToFirstTargetStep(3),
	})
}
