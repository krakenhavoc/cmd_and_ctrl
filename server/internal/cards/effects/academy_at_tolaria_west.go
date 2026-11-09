package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Academy at Tolaria West — Plane — Dominaria (#4870):
//
//	"At the beginning of your end step, if you have no cards in hand, draw
//	 seven cards.
//	 Whenever chaos ensues, discard your hand."
//
// The first ability has an intervening "if" (CR 603.4, the ruling): it
// doesn't trigger unless the planar controller's hand is empty as their end
// step begins, and does nothing unless it is still empty as it resolves. The
// chaos ability discards the whole hand (nothing, from an empty one).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "15b979de-c8ee-4664-9ca7-6c4eb3346967",
		Name:         "Academy at Tolaria West",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InterveningIf(b6YouHaveNoCardsInHand{},
				AtYourEndStep("Academy at Tolaria West — draw seven cards", Do(DrawCards{N: 7}))),
			WheneverChaosEnsues("Academy at Tolaria West — discard your hand",
				func(g *game.Game, item *game.StackItem) error {
					_, err := discardWholeHand(g, item.Controller)
					return err
				}),
		},
	})
}
