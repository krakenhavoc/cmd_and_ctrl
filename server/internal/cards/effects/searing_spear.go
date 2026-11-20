package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Searing Spear — Instant {1}{R} (Magic 2013, #5568):
//
//	"Searing Spear deals 3 damage to any target."
//
// Lightning Strike at instant speed: the same any-target clause and the
// same three damage from the spell as its source.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aafd44c1-74a9-4aa7-a4a2-67eb39a07478",
		Name:         "Searing Spear",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 3)), Answers: game.AnswerRemove},
		OnResolve:    damageToFirstTarget(3),
	})
}
