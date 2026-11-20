package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nourish — Instant {G}{G} (Darksteel, #5605):
//
//	"You gain 6 life."
//
// One GainLife step (CR 119.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9d5e82d3-79ac-4dbe-80e8-1db6dd6c8767",
		Name:         "Nourish",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{LifeGain: 6},
		OnResolve:    ResolveSteps(GainLife{Amount: 6}),
	})
}
