package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Harmonize — "Draw three cards." Same shape as Divination with a
// bigger constant; different colour + CMC but the effect is
// identical from the engine's perspective.
func init() {
	Register(Spec{
		OracleID: "7eff84f1-f772-497a-b350-bbc93d0230f7",
		Name:     "Harmonize",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return DrawCards{Player: ctx.Controller(), N: 3}.Apply(ctx)
		},
	})
}
