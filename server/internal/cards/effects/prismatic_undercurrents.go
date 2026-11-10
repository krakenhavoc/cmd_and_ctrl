package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prismatic Undercurrents — Enchantment {3}{G} (Lorwyn Eclipsed, #4458):
//
//	"Vivid — When this enchantment enters, search your library for up to X
//	 basic land cards, where X is the number of colors among permanents you
//	 control. Reveal those cards, put them into your hand, then shuffle.
//	 You may play an additional land on each of your turns."
func init() {
	Register(Spec{
		OracleID:            "27a3aa64-1de7-4b0c-af23-52886e7d258a",
		Name:                "Prismatic Undercurrents",
		Completeness:        CompletenessFull,
		AdditionalLandPlays: 1,
		Purpose:             game.Purpose{ExtraLandDrops: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Prismatic Undercurrents — search for up to X basic land cards, X the colors among permanents you control (vivid)",
				func(g *game.Game, item *game.StackItem) error {
					x := len(colorsAmongPermanentsYouControl(g, item.Controller))
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: IsBasicLand,
						Dest:      game.ZoneHand,
						Limit:     x,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Prismatic Undercurrents — up to X basic land cards",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
