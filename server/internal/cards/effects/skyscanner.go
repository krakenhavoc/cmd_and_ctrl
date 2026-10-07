package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skyscanner — Artifact Creature — Thopter {3}, 1/1:
//
//	"Flying
//	 When this creature enters, draw a card."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "974f788a-039f-4310-a2fe-16b14a1e2d35",
		Name:            "Skyscanner",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{samiDrawOnETB("Skyscanner")},
	})
}
