package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "63e5180f-16d0-4524-94bf-ecbf3c417814",
		Name:            "Yuyan Archers",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(lootLabel, func(g *game.Game, item *game.StackItem) error {
				return effects.B39MayDiscardThenDraw(1, false, lootLabel, effects.DrawOneIfDiscarded)(effects.NewContext(g, item))
			}),
		},
	})
}
