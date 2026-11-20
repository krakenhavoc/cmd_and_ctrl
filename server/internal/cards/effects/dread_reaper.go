package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dread Reaper — Creature — Horror {3}{B}{B}{B}, 6/5 (Portal, #5630):
//
//	"Flying
//	When this creature enters, you lose 5 life."
//
// Foul Imp's body and enters trigger, with 5 life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "bd73ab86-0ac9-4ce0-be41-f4ad257e74f6",
		Name:            "Dread Reaper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Dread Reaper — you lose 5 life", Do(GainLife{Amount: -5})),
		},
	})
}
