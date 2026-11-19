package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fissure Wizard — Creature — Goblin Wizard {1}{R}, 2/1:
//
//	"When this creature enters, you may discard a card. If you do, draw a
//	 card."
//
// The shared rummage of Boundary Lands Ranger: the discard is asked first and
// the draw is made only for a card actually discarded.
//
// No simplification.
func init() {
	const label = "Fissure Wizard — you may discard a card. If you do, draw a card"
	Register(Spec{
		OracleID:     "e3474632-6d81-43a8-9329-4c7e0c631755",
		Name:         "Fissure Wizard",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				return b39MayDiscardThenDraw(1, false, label, drawOneIfDiscarded)(NewContext(g, item))
			}),
		},
	})
}
