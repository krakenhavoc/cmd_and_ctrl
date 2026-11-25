package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Breath of Fire — Instant {1}{R} (Global Series Jiang Yanggu & Mu Yanling, S82 OLD batch 201, #5528):
//
//	"Breath of Fire deals 2 damage to target creature."
//
// Bathe in Dragonfire's one-creature damage at 2 and instant speed: skipped if
// the target is no longer legal as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "046923d6-2d58-42ce-afb8-886f4932e76c",
		Name:         "Breath of Fire",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose: game.Purpose{
			Answers: game.AnswerRemove,
			Targets: game.ForTargets(effects.DamageToTarget(0, 2)),
		},
		OnResolve: effects.DamageToFirstTargetStep(2),
	})
}
