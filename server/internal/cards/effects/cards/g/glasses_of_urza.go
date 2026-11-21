package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glasses of Urza — Artifact {1}:
//
//	"{T}: Look at target player's hand."
//
// "Look at" is not "reveal": only the activating player learns the
// hand (lookAtTargetPlayersHand, shared with Peek and Gitaxian Probe).
// The hand is read as the ability resolves, and a target player who
// left the game in response makes the ability fizzle (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "af7fabf4-8d55-4b06-9c21-472f4a5775b4",
		Name:         "Glasses of Urza",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Look at target player's hand",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				effects.LookAtTargetPlayersHand(item, effects.NewContext(g, item))
				return nil
			},
		}},
	})
}
