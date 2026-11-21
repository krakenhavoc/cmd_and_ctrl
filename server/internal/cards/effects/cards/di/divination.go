package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Divination — "Draw two cards." Caster draws; no announce-time
// target. Canonical "you-target" self-effect.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "273b339c-964b-4a18-8eb5-ceb8abcdfd9e",
		Name:         "Divination",
		Purpose:      game.Purpose{Draws: 2},
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: ctx.Controller(), N: 2}.Apply(ctx)
		},
	})
}
