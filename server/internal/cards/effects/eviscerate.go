package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eviscerate — Sorcery {3}{B} (Dominaria, #5530):
//
//	"Destroy target creature."
//
// "Destroy target creature." — the shared destroyTheTargetPermanent body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c3a89c67-1931-4d42-98c8-94a40cfd046e",
		Name:         "Eviscerate",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature"),
		OnResolve:    destroyTheTargetPermanent,
	})
}
