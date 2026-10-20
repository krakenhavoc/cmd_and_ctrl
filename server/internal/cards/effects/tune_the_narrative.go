package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tune the Narrative — Instant {U}:
//
//	"Draw a card. You get {E}{E} (two energy counters)."
//
// ADR 0129 PR 1.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4262a726-d1da-4bdc-ab14-73212aa83f9c",
		Name:         "Tune the Narrative",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Energy: 2},
		// Two steps, so the energy waits for a prompt the draw asks
		// (dredge, a CR 616 order), CR 608.2c (#4019).
		OnResolve: ResolveSteps(DrawCards{N: 1}, GetEnergy{N: 2}),
	})
}
