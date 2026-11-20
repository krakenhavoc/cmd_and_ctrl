package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wistful Selkie — Creature — Merfolk Wizard {G/U}{G/U}{G/U}, 2/2 (Eventide, #5590):
//
//	"When this creature enters, draw a card."
//
// Wall of Blossoms' enters trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4503d0fb-842d-49f1-9769-8509ff0f2ce1",
		Name:         "Wistful Selkie",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Wistful Selkie — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
