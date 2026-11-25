package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cleansing Screech — Sorcery {4}{R} (Global Series Jiang Yanggu & Mu Yanling, S82 OLD batch 201, #5528):
//
//	"Cleansing Screech deals 4 damage to any target."
//
// Bee Sting's any-target damage at 4: skipped if the target is no longer legal
// as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4523c22b-363b-4bc6-ba23-b428054abc5d",
		Name:         "Cleansing Screech",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 4)), Answers: game.AnswerRemove},
		OnResolve:    effects.DamageToFirstTargetStep(4),
	})
}
