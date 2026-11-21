package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spawning Breath — Instant {1}{R} (Rise of the Eldrazi, #5581):
//
//	"Spawning Breath deals 1 damage to any target. Create a 0/1 colorless
//	 Eldrazi Spawn creature token. It has "Sacrifice this token: Add {C}.""
//
// A spell with one target that has become illegal doesn't resolve at all,
// so no token is made then (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "82ce3682-0514-49a5-a330-2591e9789d36",
		Name:         "Spawning Breath",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 1)), Tokens: 1, CreatureTokens: 1, Answers: game.AnswerRemove},
		OnResolve:    effects.ResolveSteps(effects.DamageTheClauseTarget(0, 1), effects.RoeB1Spawns(1)),
	})
}
