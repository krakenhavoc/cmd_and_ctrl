package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Airborne Aid — Sorcery {3}{U} (Onslaught):
//
//	"Draw a card for each Bird on the battlefield."
//
// Timberwatch Elf's table-wide count (b42CountCreaturesWithSubtype) over
// Birds: every player's, read as the spell resolves. "Bird" is a word of
// the card's text, so it reads through a text-changing effect (CR
// 612.2). With no Bird it draws nothing. The draw is counted at
// resolution, so it declares no amount.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7f715f72-5444-4ab3-a6ca-3d5bfd0d4851",
		Name:         "Airborne Aid",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			n := effects.B42CountCreaturesWithSubtype(ctx.Game, ctx.Game.TextReaderForEffect(item).TextSubtype("Bird"))
			return effects.DrawCards{Player: item.Controller, N: n}.Apply(ctx)
		},
	})
}
