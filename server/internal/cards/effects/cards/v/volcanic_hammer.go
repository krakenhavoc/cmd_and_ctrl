package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Volcanic Hammer — Sorcery {1}{R} (Portal, #5630):
//
//	"Volcanic Hammer deals 3 damage to any target."
//
// Lightning Bolt's body as a sorcery (damageToFirstTarget). A target
// that became illegal is never dealt damage (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "98fa5a06-0553-40fd-999c-bc31c9b3f4db",
		Name:         "Volcanic Hammer",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 3)), Answers: game.AnswerRemove},
		OnResolve:    effects.DamageToFirstTargetStep(3),
	})
}
