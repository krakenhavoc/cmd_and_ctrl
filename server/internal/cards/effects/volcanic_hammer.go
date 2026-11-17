package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Volcanic Hammer — Sorcery {1}{R} (Portal, #5630):
//
//	"Volcanic Hammer deals 3 damage to any target."
//
// Lightning Bolt's body as a sorcery (damageToFirstTarget). A target
// that became illegal is never dealt damage (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "98fa5a06-0553-40fd-999c-bc31c9b3f4db",
		Name:         "Volcanic Hammer",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 3)), Answers: game.AnswerRemove},
		OnResolve:    damageToFirstTarget(3),
	})
}
