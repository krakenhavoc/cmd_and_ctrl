package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Concentrate — Sorcery {2}{U}{U} (Foundations Jumpstart):
//
//	"Draw three cards."
//
// Divination's body with three cards.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a3d287ec-1a39-4f91-acf1-7c1d2c00ed89",
		Name:         "Concentrate",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 3},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DrawCards{Player: ctx.Controller(), N: 3}.Apply(ctx)
		},
	})
}
