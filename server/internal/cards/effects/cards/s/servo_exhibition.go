package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Servo Exhibition — Sorcery {1}{W} (Kaladesh, #5539):
//
//	"Create two 1/1 colorless Servo artifact creature tokens."
//
// Two Servos from the token table, the same token fabricate makes.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "76a6ddc2-f894-4331-89a9-87ca16efd0c0",
		Name:         "Servo Exhibition",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2},
		OnResolve:    effects.ResolveSteps(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Servo artifact"), N: 2}),
	})
}
