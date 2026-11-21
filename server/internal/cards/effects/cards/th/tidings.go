package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tidings — Sorcery {3}{U}{U} (Starter 1999, #5620):
//
//	"Draw four cards."
//
// Into the Story's draw without its cost reduction.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "72897780-094d-4a21-8b1c-419a9defd2fb",
		Name:         "Tidings",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 4},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: ctx.Controller(), N: 4}.Apply(ctx)
		},
	})
}
