package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Avenging Hunter — Creature — Dragon Ranger {4}{G}, 5/4:
//
//	"Trample
//	 When this creature enters, you take the initiative."
//
// Trample is engine-side; the ETB is TakeTheInitiative (ADR 0203,
// #4096).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e0aa5e3d-17c0-4ac7-9d76-9ce579d1eb2b",
		Name:            "Avenging Hunter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouTakeTheInitiative("Avenging Hunter"),
		},
	})
}
