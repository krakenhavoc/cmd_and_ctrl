package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Casey Jones, Vigilante — Legendary Creature — Human Berserker {1}{R}{R},
// 4/3 (Teenage Mutant Ninja Turtles, #4457):
//
//	"When Casey Jones enters, draw three cards. At the beginning of your
//	 next upkeep, discard three cards at random."
//
// One enters trigger whose resolution draws three and schedules a CR
// 603.7 delayed trigger for the controller's next upkeep (Pact of
// Negation's ControllerTurnOnly schedule). The delayed trigger is the
// controller's whether or not Casey Jones is still around, and the
// discard is at random from the hand as it is then (CR 701.9b): a hand
// of fewer than three cards is discarded whole.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bc163b5a-0b1a-4257-8c07-5c7d5284015f",
		Name:         "Casey Jones, Vigilante",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Casey Jones, Vigilante — draw three cards; discard three at random at your next upkeep",
				Do(DrawCards{N: 3}, ScheduleDelayedTrigger{
					At:                 game.StepUpkeep,
					ControllerTurnOnly: true,
					Label:              "Casey Jones, Vigilante — discard three cards at random",
					Body:               caseyJonesDiscardBody,
				})),
		},
	})
}

// caseyJonesDiscardBody is the delayed trigger: its controller discards
// three cards at random.
var caseyJonesDiscardBody = game.SimpleDelayedBody("casey-jones-vigilante/discard-three-at-random",
	func(g *game.Game, item *game.StackItem) error {
		return DiscardCards{Player: item.Controller, N: 3}.Apply(NewContext(g, item))
	})
