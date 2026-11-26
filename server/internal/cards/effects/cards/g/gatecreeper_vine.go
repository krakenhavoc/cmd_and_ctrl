package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gatecreeper Vine — Creature — Plant {1}{G}, 0/2 (Return to Ravnica, S82
// OLD):
//
//	"Defender
//	 When this creature enters, you may search your library for a basic land
//	 card or a Gate card, reveal it, put it into your hand, then shuffle."
//
// Outcaster Greenblade's search with Gate for Desert: a basic land card or a
// land with the Gate subtype, revealed and put into your hand. "You may
// search" lets you decline the search and its shuffle.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0f539127-535f-4e0d-abaa-e884521098d2",
		Name:            "Gatecreeper Vine",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Gatecreeper Vine — you may search your library for a basic land card or a Gate card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player: item.Controller,
						Predicate: func(c game.Card) bool {
							return effects.B30IsBasicLandCard(c) || (c.IsLand() && c.HasSubtype("Gate"))
						},
						Dest:     game.ZoneHand,
						Limit:    1,
						Reveal:   true,
						Shuffle:  true,
						Optional: true,
						Reason:   "Gatecreeper Vine — a basic land card or a Gate card",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
