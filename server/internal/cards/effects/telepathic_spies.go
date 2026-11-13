package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "8162730e-2573-4bc0-827f-5b7b3dbf3bff",
		Name:         "Telepathic Spies",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Telepathic Spies — look at target opponent's hand",
				func(g *game.Game, item *game.StackItem) error {
					lookAtTargetPlayersHand(item, NewContext(g, item))
					return nil
				}),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}
