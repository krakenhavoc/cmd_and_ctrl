package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wanderguard Sentry — Creature — Drone {4}{U}, 3/3 (Mirrodin, #5606):
//
//	"When this creature enters, look at target opponent's hand."
//
// Telepathic Spies' enters trigger: the hand is shown to the trigger's
// controller alone, and nothing else changes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cd02e276-3c2d-4177-aa60-294a31bae0db",
		Name:         "Wanderguard Sentry",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Wanderguard Sentry — look at target opponent's hand",
				func(g *game.Game, item *game.StackItem) error {
					lookAtTargetPlayersHand(item, NewContext(g, item))
					return nil
				}),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}
