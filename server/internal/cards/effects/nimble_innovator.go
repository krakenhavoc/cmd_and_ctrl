package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nimble Innovator — Creature — Vedalken Artificer {3}{U}, 2/2 (Kaladesh, #5539):
//
//	"When this creature enters, draw a card."
//
// Cloudkin Seer's enters draw, on the stack so it can be answered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "30929713-47dc-40f0-883c-4bd41aa399d5",
		Name:         "Nimble Innovator",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Nimble Innovator — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
