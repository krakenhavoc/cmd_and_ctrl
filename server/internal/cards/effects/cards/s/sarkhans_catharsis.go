package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sarkhan's Catharsis — Instant {4}{R} (War of the Spark, #5520):
//
//	"Sarkhan's Catharsis deals 5 damage to target player or planeswalker."
//
// Lava Axe's body, at instant speed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "787935ab-26a6-4b4a-9c14-e106160e951d",
		Name:         "Sarkhan's Catharsis",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayerOrPlaneswalker(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 5)), Answers: game.AnswerRemove},
		OnResolve:    effects.DamageToFirstTargetStep(5),
	})
}
