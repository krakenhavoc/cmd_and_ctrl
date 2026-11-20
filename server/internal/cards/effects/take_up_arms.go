package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Take Up Arms — Instant {4}{W} (Khans of Tarkir, #5554):
//
//	"Create three 1/1 white Warrior creature tokens."
//
// Three 1/1 white Warriors from the token table, made at once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "61e2d557-f08d-489b-94c8-aff87e5189f9",
		Name:         "Take Up Arms",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 3, CreatureTokens: 3},
		OnResolve:    ResolveSteps(CreateToken{Template: TokenCard("1/1 white Warrior"), N: 3}),
	})
}
