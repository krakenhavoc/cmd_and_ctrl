package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tune the Narrative — Instant {U}:
//
//	"Draw a card. You get {E}{E} (two energy counters)."
//
// ADR 0129 PR 1.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4262a726-d1da-4bdc-ab14-73212aa83f9c",
		Name:         "Tune the Narrative",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Energy: 2},
		// Two steps, so the energy waits for a prompt the draw asks
		// (dredge, a CR 616 order), CR 608.2c (#4019).
		OnResolve: effects.ResolveSteps(effects.DrawCards{N: 1}, effects.GetEnergy{N: 2}),
	})
}
