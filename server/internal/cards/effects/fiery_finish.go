package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fiery Finish — Sorcery {4}{R}{R} (Core Set 2019, #5527):
//
//	"Fiery Finish deals 7 damage to target creature."
//
// Bathe in Dragonfire's text with 7 damage: the shared
// damageToFirstTarget.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "afa3eb61-624c-4abe-bf1e-58fa3b4c14a6",
		Name:         "Fiery Finish",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose: game.Purpose{
			Answers: game.AnswerRemove,
			Targets: game.ForTargets(DamageToTarget(0, 7)),
		},
		OnResolve: damageToFirstTarget(7),
	})
}
