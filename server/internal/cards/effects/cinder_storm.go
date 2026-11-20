package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cinder Storm — Sorcery {6}{R} (Starter 1999, #5620):
//
//	"Cinder Storm deals 7 damage to any target."
//
// Volcanic Hammer's body at 7 (damageToFirstTarget). A target that
// became illegal is never dealt damage (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c7971968-c4ee-4437-92ec-1238bce1935f",
		Name:         "Cinder Storm",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 7)), Answers: game.AnswerRemove},
		OnResolve:    damageToFirstTarget(7),
	})
}
