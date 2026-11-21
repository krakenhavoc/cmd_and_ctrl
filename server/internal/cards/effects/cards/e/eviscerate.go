package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eviscerate — Sorcery {3}{B} (Dominaria, #5530):
//
//	"Destroy target creature."
//
// "Destroy target creature." — the shared destroyTheTargetPermanent body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c3a89c67-1931-4d42-98c8-94a40cfd046e",
		Name:         "Eviscerate",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
