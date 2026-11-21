package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cinder Storm — Sorcery {6}{R} (Starter 1999, #5620):
//
//	"Cinder Storm deals 7 damage to any target."
//
// Volcanic Hammer's body at 7 (damageToFirstTarget). A target that
// became illegal is never dealt damage (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c7971968-c4ee-4437-92ec-1238bce1935f",
		Name:         "Cinder Storm",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 7)), Answers: game.AnswerRemove},
		OnResolve:    effects.DamageToFirstTargetStep(7),
	})
}
