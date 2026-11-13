package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Counsel of the Soratami — Sorcery {2}{U} (Champions of Kamigawa,
// #5603):
//
//	"Draw two cards."
//
// Divination.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "62ddc5ae-ced9-4319-854c-1a114c6afc3f",
		Name:         "Counsel of the Soratami",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DrawCards{Player: ctx.Controller(), N: 2}.Apply(ctx)
		},
	})
}
