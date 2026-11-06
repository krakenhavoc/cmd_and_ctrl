package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pond Prophet — Creature — Frog Advisor {G/U}{G/U}, 1/1 (Bloomburrow,
// #4467):
//
//	"When this creature enters, draw a card."
//
// The shared enters trigger (Elvish Visionary's shape): it uses the
// stack, so it can be answered, and fires however the creature enters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6eabebdf-c1a3-4789-a67f-859f0957c5df",
		Name:         "Pond Prophet",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Pond Prophet — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
