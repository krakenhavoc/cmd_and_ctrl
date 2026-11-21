package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Return from the Wilds — Sorcery {2}{G} (Wilds of Eldraine, #4471):
//
//	"Choose two —
//	 • Search your library for a basic land card, put it onto the
//	   battlefield tapped, then shuffle.
//	 • Create a 1/1 white Human creature token.
//	 • Create a Food token."
//
// Atarka's Command's "choose two" (CR 700.2): the two modes are announced
// with the cast and resolve in printed order (CR 700.2a). The search is
// Reconnaissance Partition's basic land onto the battlefield tapped; a
// library with no basic still shuffles. The Human is the token table's
// 1/1 white Human and the Food the catalog's Food.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "541adb13-1277-4c9e-ae1e-976f7aac1c72",
		Name:         "Return from the Wilds",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose two", 2, 2,
			effects.ModeDoing("Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.", nil,
				func(item *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.FetchBasicTapped(ctx.Game, item)
				}),
			effects.ModeDoing("Create a 1/1 white Human creature token.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.CreateToken{Controller: ctx.Controller(), Template: effects.TokenCard("1/1 white Human"), N: 1}.Apply(ctx)
				}),
			effects.ModeDoing("Create a Food token.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.CreateToken{Controller: ctx.Controller(), Template: effects.FoodToken(), N: 1}.Apply(ctx)
				}),
		),
	})
}
