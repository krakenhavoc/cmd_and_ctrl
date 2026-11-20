package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Resupply — Instant {5}{W} (Dragons of Tarkir,
// #5551):
//
//	"You gain 6 life.
//	 Draw a card."
//
// Dosan's Oldest Chant's two instructions, in printed order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "129ec223-e76d-491d-b299-07ac91168d31",
		Name:         "Resupply",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{LifeGain: 6, Draws: 1},
		OnResolve:    ResolveSteps(GainLife{Amount: 6}, DrawCards{N: 1}),
	})
}
