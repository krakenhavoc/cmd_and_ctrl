package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Divination — "Draw two cards." Caster draws; no announce-time
// target. Canonical "you-target" self-effect.
func init() {
	Register(Spec{
		OracleID:     "273b339c-964b-4a18-8eb5-ceb8abcdfd9e",
		Name:         "Divination",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return DrawCards{Player: ctx.Controller(), N: 2}.Apply(ctx)
		},
	})
}
