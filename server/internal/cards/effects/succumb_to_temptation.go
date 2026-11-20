package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Succumb to Temptation — Instant {1}{B}{B} (Eldritch Moon, #5541):
//
//	"You draw two cards and you lose 2 life."
//
// Night's Whisper's text at instant speed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c8e21bc4-ea57-46d6-9e7e-ad95e17ea0fd",
		Name:         "Succumb to Temptation",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		OnResolve: ResolveSteps(
			// The life loss waits for a prompt the draw asks (CR 608.2c, #4155).
			DrawCards{N: 2},
			GainLife{Amount: -2},
		),
	})
}
