package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Explosive Impact — Instant {5}{R} (Return to Ravnica, #5566):
//
//	"Explosive Impact deals 5 damage to any target."
//
// Bee Sting's body at five: the shared first-target damage, re-checked at
// resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5eae77d0-7653-4b07-a701-aabea13991bf",
		Name:         "Explosive Impact",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 5)), Answers: game.AnswerRemove},
		OnResolve:    damageToFirstTarget(5),
	})
}
