package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Secret Plans — Enchantment {G}{U} (Commander 2019, #5517):
//
//	"Face-down creatures you control get +0/+1.
//	 Whenever a permanent you control is turned face up, draw a card."
//
// The static is a layer 7c anthem over the face-down creatures its controller
// controls (a face-down permanent is a 2/2 creature, CR 708.2a); the trigger
// watches the turned-face-up event for any permanent you control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7d0d7480-fddf-4841-bc2a-239e03f97f60",
		Name:         "Secret Plans",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.B16Anthem(effects.OldB153FaceDownCreaturesYouControl, 0, 1)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventTurnedFaceUp, effects.OldB153PermanentYouControlTurnedFaceUp,
				"Secret Plans — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
