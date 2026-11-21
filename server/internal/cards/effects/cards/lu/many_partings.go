package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Many Partings — Sorcery {G} (EDHREC rank 1957):
//
//	"Search your library for a basic land card, reveal it, put it
//	 into your hand, then shuffle. Create a Food token."
//
// A one-mana Lay of the Land that leaves a Food behind. The search
// is the searcher's prompt when the library holds more than one
// basic; the Food is the shared artifact-token template. The Food
// is created in the search's continuation so it lands after the
// land is in hand, in the printed order, whether or not a prompt
// was needed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "92ea750e-62b7-4422-a527-ddf435759d08",
		Name:         "Many Partings",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:    item.Controller,
				Predicate: effects.IsBasicLand,
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Many Partings — a basic land card, revealed, to hand",
				ThenTail:  ctx.ItemTail(manyPartingsFood),
			}.Apply(ctx)
		},
	})
}

// manyPartingsFood is "create a Food token". An on-disk key.
var manyPartingsFood = effects.CardTail("many-partings/create-food", func(ctx *effects.Context, _ game.TailOutcome) error {
	return effects.CreateToken{Controller: ctx.Controller(), Template: effects.FoodToken(), N: 1}.Apply(ctx)
})
