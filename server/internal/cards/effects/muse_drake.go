package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Muse Drake — Creature — Drake {3}{U}, 1/3 (Foundations Jumpstart):
//
//	"Flying
//	 When this creature enters, draw a card."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "04c9b3d7-83ba-400e-9837-2ba0119e9445",
		Name:            "Muse Drake",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Muse Drake — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
