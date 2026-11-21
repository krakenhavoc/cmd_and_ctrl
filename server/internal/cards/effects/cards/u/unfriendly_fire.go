package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unfriendly Fire — Instant {4}{R} (Ixalan, #5532):
//
//	"Unfriendly Fire deals 4 damage to any target."
//
// Flame Lash's body: 4 damage to the target, nothing if it has become
// illegal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ebb5b792-3906-4cb5-b8f9-063f94c9d135",
		Name:         "Unfriendly Fire",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 4)), Answers: game.AnswerRemove},
		OnResolve:    effects.DamageToFirstTargetStep(4),
	})
}
