package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scorching Shot — Sorcery {R}{R} (Outlaws of Thunder Junction, #4468):
//
//	"Scorching Shot deals 5 damage to target creature."
//
// Lightning Bolt's body (damageToFirstTarget) at 5, to a creature only.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e13113ba-1923-493c-a48d-ca320b7adbde",
		Name:         "Scorching Shot",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 5)), Answers: game.AnswerRemove},
		OnResolve:    damageToFirstTarget(5),
	})
}
