package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Telepathic Spies — Creature — Human Wizard {2}{U}, 2/2 (Urza's
// Destiny, #5621):
//
//	"When this creature enters, look at target opponent's hand."
//
// Glasses of Urza's look as an enters trigger aimed at an opponent:
// "look at" is not "reveal", so only the Spies' controller learns the
// hand (lookAtTargetPlayersHand). The hand is read as the trigger
// resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8162730e-2573-4bc0-827f-5b7b3dbf3bff",
		Name:         "Telepathic Spies",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Telepathic Spies — look at target opponent's hand",
				func(g *game.Game, item *game.StackItem) error {
					effects.LookAtTargetPlayersHand(item, effects.NewContext(g, item))
					return nil
				}),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
