package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dawning Angel — Creature — Angel {4}{W}, 3/2 (Core Set 2020, #5518):
//
//	"Flying
//	 When this creature enters, you gain 4 life."
//
// Angel of Mercy's shape: the enters trigger uses the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a7d88510-d2ce-4dd6-938c-964cc74ff7d7",
		Name:            "Dawning Angel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Dawning Angel — you gain 4 life", Do(GainLife{Amount: 4})),
		},
	})
}
