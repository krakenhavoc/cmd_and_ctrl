package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jace's Ingenuity — Instant {3}{U}{U} (Magic 2011, #5579):
//
//	"Draw three cards."
//
// Divination's draw at three.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ce588fd3-642e-4df9-8061-6fcfc6e81270",
		Name:         "Jace's Ingenuity",
		Purpose:      game.Purpose{Draws: 3},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DrawCards{Player: ctx.Controller(), N: 3}.Apply(ctx)
		},
	})
}
