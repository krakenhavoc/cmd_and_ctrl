package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dragon's Presence — Instant {2}{W} (Global Series Jiang Yanggu & Mu Yanling, S82 OLD batch 201, #5528):
//
//	"Dragon's Presence deals 5 damage to target attacking or blocking
//	 creature."
//
// Soldier Replica's attacking-or-blocking clause with Bathe in Dragonfire's
// damage at 5: the target must be in combat as it is chosen and again as the
// spell resolves, or the spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4fb7aeb6-2fc3-459e-bad8-1c7ea92f7ae7",
		Name:         "Dragon's Presence",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
		Purpose: game.Purpose{
			Answers: game.AnswerRemove,
			Targets: game.ForTargets(effects.DamageToTarget(0, 5)),
		},
		OnResolve: effects.DamageToFirstTargetStep(5),
	})
}
