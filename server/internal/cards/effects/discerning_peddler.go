package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Discerning Peddler — Creature — Human Rogue {1}{R}, 2/2 (Outlaws of
// Thunder Junction, #4468):
//
//	"When this creature enters, you may discard a card. If you do, draw a
//	 card."
//
// Rubble Rouser's enters loot: the shared discard-then-draw prompt, where
// the draw follows only a discard that happened. Declining, or an empty hand,
// draws nothing.
//
// No simplification.
func init() {
	const lootLabel = "Discerning Peddler — you may discard a card; if you do, draw a card"
	Register(Spec{
		OracleID:     "a6584af7-583d-4d89-9b46-d61be8236413",
		Name:         "Discerning Peddler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters(lootLabel, func(g *game.Game, item *game.StackItem) error {
				return b39MayDiscardThenDraw(1, false, lootLabel, drawOneIfDiscarded)(NewContext(g, item))
			}),
		},
	})
}
