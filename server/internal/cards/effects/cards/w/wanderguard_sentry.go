package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wanderguard Sentry — Creature — Drone {4}{U}, 3/3 (Mirrodin, #5606):
//
//	"When this creature enters, look at target opponent's hand."
//
// Telepathic Spies' enters trigger: the hand is shown to the trigger's
// controller alone, and nothing else changes.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cd02e276-3c2d-4177-aa60-294a31bae0db",
		Name:         "Wanderguard Sentry",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Wanderguard Sentry — look at target opponent's hand",
				func(g *game.Game, item *game.StackItem) error {
					effects.LookAtTargetPlayersHand(item, effects.NewContext(g, item))
					return nil
				}),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
