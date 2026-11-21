package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prismatic Undercurrents — Enchantment {3}{G} (Lorwyn Eclipsed, #4458):
//
//	"Vivid — When this enchantment enters, search your library for up to X
//	 basic land cards, where X is the number of colors among permanents you
//	 control. Reveal those cards, put them into your hand, then shuffle.
//	 You may play an additional land on each of your turns."
func init() {
	effects.Register(effects.Spec{
		OracleID:            "27a3aa64-1de7-4b0c-af23-52886e7d258a",
		Name:                "Prismatic Undercurrents",
		Completeness:        effects.CompletenessFull,
		AdditionalLandPlays: 1,
		Purpose:             game.Purpose{ExtraLandDrops: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Prismatic Undercurrents — search for up to X basic land cards, X the colors among permanents you control (vivid)",
				func(g *game.Game, item *game.StackItem) error {
					x := len(effects.ColorsAmongPermanentsYouControl(g, item.Controller))
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: effects.IsBasicLand,
						Dest:      game.ZoneHand,
						Limit:     x,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Prismatic Undercurrents — up to X basic land cards",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
