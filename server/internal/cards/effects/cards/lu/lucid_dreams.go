package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lucid Dreams — Sorcery {3}{U}{U} (Modern Horizons 2, #5501):
//
//	"Draw X cards, where X is the number of card types among cards in your
//	 graveyard."
//
// X is counted as the spell resolves, so Lucid Dreams itself (still on the
// stack) is not in the graveyard to count. A card with several types counts
// once for each (CR 205.2a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bd716715-15a4-4356-ad7e-c642fafbfd18",
		Name:         "Lucid Dreams",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: ctx.Controller(), N: effects.B16CardTypesInGraveyard(ctx.Game, ctx.Controller())}.Apply(ctx)
		},
	})
}
