package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goliath Paladin — Creature — Giant Knight {4}{W}, 3/6:
//
//	"Vigilance
//	 When this creature enters, you take the initiative."
//
// Vigilance is engine-side; the ETB is TakeTheInitiative (ADR 0203,
// #4096).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "269a02ca-1425-4b32-81f3-440e1caadb17",
		Name:            "Goliath Paladin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouTakeTheInitiative("Goliath Paladin"),
		},
	})
}
