package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Yuyan Archers — Creature — Human Archer {1}{R}, 3/1 (Avatar: The Last
// Airbender, #4459):
//
//	"Reach
//	 When this creature enters, you may discard a card. If you do, draw
//	 a card."
//
// Rubble Rouser's entry loot: the discard is chosen as the trigger
// resolves and may be declined; only a card actually discarded draws
// one, so an empty hand draws nothing.
//
// No simplification.
func init() {
	const lootLabel = "Yuyan Archers — you may discard a card; if you do, draw a card"
	Register(Spec{
		OracleID:        "63e5180f-16d0-4524-94bf-ecbf3c417814",
		Name:            "Yuyan Archers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters(lootLabel, func(g *game.Game, item *game.StackItem) error {
				return b39MayDiscardThenDraw(1, false, lootLabel, drawOneIfDiscarded)(NewContext(g, item))
			}),
		},
	})
}
