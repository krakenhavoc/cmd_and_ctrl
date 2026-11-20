package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Weave Fate — Instant {3}{U} (Khans of Tarkir, #5554):
//
//	"Draw two cards."
//
// Touch of Brilliance's body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ff568a29-e31a-4bc3-b8a3-8fa56f91d52d",
		Name:         "Weave Fate",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		OnResolve:    ResolveSteps(DrawCards{N: 2}),
	})
}
