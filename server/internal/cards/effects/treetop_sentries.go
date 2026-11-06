package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "5bf5739a-6322-4e6e-b41e-e7bdbb11b1ce",
		Name:            "Treetop Sentries",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Treetop Sentries — you may forage; if you do, draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return MayForage(NewContext(g, item), "Treetop Sentries — forage to draw a card?", mayPayDrawACard)
				}),
		},
	})
}
