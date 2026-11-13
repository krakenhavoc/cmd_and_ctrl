package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "7f715f72-5444-4ab3-a6ca-3d5bfd0d4851",
		Name:         "Airborne Aid",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			n := b42CountCreaturesWithSubtype(ctx.Game, ctx.Game.TextReaderForEffect(item).TextSubtype("Bird"))
			return DrawCards{Player: item.Controller, N: n}.Apply(ctx)
		},
	})
}
