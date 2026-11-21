package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Concentrate — Sorcery {2}{U}{U} (Foundations Jumpstart):
//
//	"Draw three cards."
//
// Divination's body with three cards.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a3d287ec-1a39-4f91-acf1-7c1d2c00ed89",
		Name:         "Concentrate",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 3},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: ctx.Controller(), N: 3}.Apply(ctx)
		},
	})
}
