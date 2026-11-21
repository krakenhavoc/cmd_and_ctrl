package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hamato Ninpō — Instant {1}{W} (Teenage Mutant Ninja Turtles Commander,
// #5011):
//
//	"Hamato Ninpō deals 4 damage to target attacking or blocking
//	 creature."
//
// Eiganjo's target clause (AttackingOrBlocking reads the combat
// designations, so it is legal only during a combat with attackers or
// blockers) and the spell as the damage source. A target no longer
// attacking or blocking by resolution takes nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0587fe46-1bed-4ef3-b087-04d4cc6a334d",
		Name:         "Hamato Ninpō",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 4)), Answers: game.AnswerRemove},
		OnResolve:    effects.DamageToFirstTargetStep(4),
	})
}
