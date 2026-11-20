package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Azra Bladeseeker — Creature — Azra Warrior {2}{R}, 3/2 (Battlebond,
// #5529):
//
//	"When this creature enters, each player on your team may discard a
//	 card, then each player who discarded a card this way draws a card."
//
// "Your team" is "you" (CR 102.4), so the only player on it is the
// controller: Rubble Rouser's loot, the shared discard-then-draw prompt
// (b39MayDiscardThenDraw), where the draw only follows a discard that
// happened.
//
// No simplification.
func init() {
	const label = "Azra Bladeseeker — you may discard a card; if you do, draw a card"
	Register(Spec{
		OracleID:     "ac4d065e-7eb6-4224-8a7c-e57f2c1a2733",
		Name:         "Azra Bladeseeker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				return b39MayDiscardThenDraw(1, false, label, drawOneIfDiscarded)(NewContext(g, item))
			}),
		},
	})
}
