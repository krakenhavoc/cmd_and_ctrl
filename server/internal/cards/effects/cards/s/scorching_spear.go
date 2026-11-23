package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scorching Spear — Sorcery {R} (Portal, #5630):
//
//	"Scorching Spear deals 1 damage to any target."
//
// Gut Shot's body as a sorcery (damageToFirstTarget). A target that
// became illegal is never dealt damage (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9342fbb8-ab35-4895-946d-951ba6a2b067",
		Name:         "Scorching Spear",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 1)), Answers: game.AnswerRemove},
		OnResolve:    effects.DamageToFirstTargetStep(1),
	})
}
