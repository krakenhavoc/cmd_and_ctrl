package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tarfire — Kindred Instant — Goblin {R} (Lorwyn, #5593):
//
//	"Tarfire deals 2 damage to any target."
//
// Shock as a Kindred Goblin card (CR 308.1): cast and resolved as an
// instant, and a Goblin card for anything that counts them in a hand or a
// graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f07b821c-bbc1-4def-b009-6c03689478d0",
		Name:         "Tarfire",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 2)), Answers: game.AnswerRemove},
		OnResolve:    effects.DamageToFirstTargetStep(2),
	})
}
