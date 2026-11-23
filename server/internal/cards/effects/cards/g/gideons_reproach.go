package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gideon's Reproach — Instant {1}{W} (Battle for Zendikar, #5547):
//
//	"Gideon's Reproach deals 4 damage to target attacking or
//	 blocking creature."
//
// Divine Arrow's sentence: only a creature that is attacking or
// blocking may be the target, and it is checked again at resolution (CR
// 608.2b), so a creature removed from combat in response is not dealt
// the damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ec43d67c-2254-4782-8bd4-3318b25907e6",
		Name:         "Gideon's Reproach",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 4)), Answers: game.AnswerRemove},
		OnResolve:    effects.DamageToFirstTargetStep(4),
	})
}
