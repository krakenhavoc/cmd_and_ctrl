package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Viashino Racketeer — Creature — Lizard Rogue {2}{R}, 2/1 (Return to
// Ravnica, #5566):
//
//	"When this creature enters, you may discard a card. If you do, draw a
//	 card."
//
// Fissure Wizard's enters loot: the shared discard-then-draw prompt, where
// the draw follows only a discard that happened. Declining, or an empty
// hand, draws nothing.
//
// No simplification.
func init() {
	const racketeerLabel = "Viashino Racketeer — you may discard a card. If you do, draw a card"
	Register(Spec{
		OracleID:     "18b09c2d-20cb-40e9-8209-8a1ec2edde97",
		Name:         "Viashino Racketeer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters(racketeerLabel, func(g *game.Game, item *game.StackItem) error {
				return b39MayDiscardThenDraw(1, false, racketeerLabel, drawOneIfDiscarded)(NewContext(g, item))
			}),
		},
	})
}
