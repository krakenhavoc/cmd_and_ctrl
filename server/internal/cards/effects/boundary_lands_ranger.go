package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Boundary Lands Ranger — Creature — Human Ranger {1}{R}, 2/2 (Wilds of
// Eldraine, #4471):
//
//	"At the beginning of combat on your turn, if you control a creature
//	 with power 4 or greater, you may discard a card. If you do, draw a
//	 card."
//
// An intervening if (CR 603.4): the trigger looks for a creature you
// control with power 4 or more (the Ranger counts if it has grown) as
// your beginning of combat starts, and again as it resolves, so a board
// that lost its big creature in response does nothing. The rummage is
// Rubble Rouser's: the discard first, and a draw only for a card
// discarded.
//
// No simplification.
func init() {
	const label = "Boundary Lands Ranger — you may discard a card; if you do, draw a card"
	Register(Spec{
		OracleID:     "1189b5d5-ca8f-4b4e-a5f1-91122a485cf3",
		Name:         "Boundary Lands Ranger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondPowerAtLeast{N: 4},
				On(game.EventStepBegan, AllOf(StepBegan(game.StepBeginCombat, true),
					func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
						return b24ControlsCreatureWithPowerAtLeast(g, source.Controller, 4)
					}),
					label, func(g *game.Game, item *game.StackItem) error {
						return b39MayDiscardThenDraw(1, false, label, drawOneIfDiscarded)(NewContext(g, item))
					})),
		},
	})
}
