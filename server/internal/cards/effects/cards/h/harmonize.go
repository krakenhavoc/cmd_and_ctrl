package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Harmonize — "Draw three cards." Same shape as Divination with a
// bigger constant; different colour + CMC but the effect is
// identical from the engine's perspective.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7eff84f1-f772-497a-b350-bbc93d0230f7",
		Name:         "Harmonize",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 3},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: ctx.Controller(), N: 3}.Apply(ctx)
		},
	})
}
