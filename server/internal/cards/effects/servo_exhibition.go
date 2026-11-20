package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Servo Exhibition — Sorcery {1}{W} (Kaladesh, #5539):
//
//	"Create two 1/1 colorless Servo artifact creature tokens."
//
// Two Servos from the token table, the same token fabricate makes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "76a6ddc2-f894-4331-89a9-87ca16efd0c0",
		Name:         "Servo Exhibition",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2},
		OnResolve:    ResolveSteps(CreateToken{Template: TokenCard("1/1 colorless Servo artifact"), N: 2}),
	})
}
