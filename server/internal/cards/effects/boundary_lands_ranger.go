package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Boundary Lands Ranger — Creature {1}{R}, 2/3 Human Ranger (Wilds of
// Eldraine, #4471):
//
//	"At the beginning of combat on your turn, if you control a creature
//	 with power 4 or greater, you may discard a card. If you do, draw a
//	 card."
//
// The power-4 test is an intervening "if" (CR 603.4): checked as combat
// begins and again as the ability resolves, so a trigger whose condition
// stopped holding does nothing and asks nothing. The rummage is the shared
// "you may discard a card. If you do, draw a card" (Ashling, Rekindled).
//
// No simplification.
func init() {
	const label = "Boundary Lands Ranger — you may discard a card. If you do, draw a card"
	Register(Spec{
		OracleID:     "1189b5d5-ca8f-4b4e-a5f1-91122a485cf3",
		Name:         "Boundary Lands Ranger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondPowerAtLeast{N: 4}, AtBeginningOfYourCombat(label,
				func(g *game.Game, item *game.StackItem) error {
					return b39MayDiscardThenDraw(1, false, label, drawOneIfDiscarded)(NewContext(g, item))
				})),
		},
	})
}
