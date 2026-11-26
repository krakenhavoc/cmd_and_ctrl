package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hysterical Blindness — Instant {2}{U} (Innistrad, S82 older sets batch
// 449, #5572):
//
//	"Creatures your opponents control get -4/-0 until end of turn."
//
// Blinding Spray's boost without the draw: the set is locked at resolution
// (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "705531b3-77cc-4d1f-851b-45e03c029991",
		Name:         "Hysterical Blindness",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.BoostUntilEOT{
				Match: effects.And(effects.Creature(), effects.OpponentControls()), Power: -4,
				Label: "Hysterical Blindness — creatures your opponents control get -4/-0",
			},
		),
	})
}
