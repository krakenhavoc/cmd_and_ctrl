package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Treetop Sentries — Creature — Squirrel Archer {3}{G}, 2/4 (Bloomburrow,
// #4467):
//
//	"Reach
//	 When this creature enters, you may forage. If you do, draw a card.
//	 (To forage, exile three cards from your graveyard or sacrifice a
//	 Food.)"
//
// "You may forage. If you do" is MayForage (#5184): one prompt naming
// the way, offered only when the controller has three cards in their
// graveyard or a Food, and the draw happens once the forage is paid.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5bf5739a-6322-4e6e-b41e-e7bdbb11b1ce",
		Name:            "Treetop Sentries",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Treetop Sentries — you may forage; if you do, draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayForage(effects.NewContext(g, item), "Treetop Sentries — forage to draw a card?", effects.MayPayDrawACard)
				}),
		},
	})
}
