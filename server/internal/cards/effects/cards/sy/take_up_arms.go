package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Take Up Arms — Instant {4}{W} (Khans of Tarkir, #5554):
//
//	"Create three 1/1 white Warrior creature tokens."
//
// Three 1/1 white Warriors from the token table, made at once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "61e2d557-f08d-489b-94c8-aff87e5189f9",
		Name:         "Take Up Arms",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 3, CreatureTokens: 3},
		OnResolve:    effects.ResolveSteps(effects.CreateToken{Template: effects.TokenCard("1/1 white Warrior"), N: 3}),
	})
}
