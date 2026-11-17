package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scorching Spear — Sorcery {R} (Portal, #5630):
//
//	"Scorching Spear deals 1 damage to any target."
//
// Gut Shot's body as a sorcery (damageToFirstTarget). A target that
// became illegal is never dealt damage (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9342fbb8-ab35-4895-946d-951ba6a2b067",
		Name:         "Scorching Spear",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 1)), Answers: game.AnswerRemove},
		OnResolve:    damageToFirstTarget(1),
	})
}
