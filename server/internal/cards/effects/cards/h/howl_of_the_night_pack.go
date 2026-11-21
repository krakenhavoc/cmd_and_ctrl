package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Howl of the Night Pack — Sorcery {6}{G} (Shadowmoor, #5591):
//
//	"Create a 2/2 green Wolf creature token for each Forest you
//	 control."
//
// Gather the White Lotus' count (a land subtype you control, read as the
// spell resolves through its text, CR 612.2) feeding the table's 2/2 green
// Wolf.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "868b758c-e21c-4079-a1fc-4c89af106f4b",
		Name:         "Howl of the Night Pack",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			n := effects.B33IslandsControlled(ctx.Game, ctx.Controller(), ctx.TextSubtype("Forest"))
			if n <= 0 {
				return nil
			}
			return effects.CreateToken{Template: effects.TokenCard("2/2 green Wolf"), N: n}.Apply(ctx)
		},
	})
}
