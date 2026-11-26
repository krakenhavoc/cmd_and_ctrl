package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lair Delve — {2}{G} Sorcery: "Reveal the top two cards of your library. Put all
// creature and land cards revealed this way into your hand and the rest on the
// bottom of your library in any order."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8d7ce7b5-2e56-4060-9406-f700ce1fbb7d",
		Name:         "Lair Delve",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.LairDelveBody(ctx.Game, item)
		},
	})
}
