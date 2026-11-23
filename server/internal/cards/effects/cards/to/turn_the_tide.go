package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Turn the Tide — Instant {1}{U} (Mirrodin Besieged,
// #5576):
//
//	"Creatures your opponents control get -2/-0 until end of turn."
//
// Uncomfortable Chill's shrink without the draw: the set of creatures your
// opponents control is fixed as it resolves (CR 611.2c), so a creature that
// enters later is not shrunk.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2546db90-1f6e-43f7-a116-2fefe67d489d",
		Name:         "Turn the Tide",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.BoostUntilEOT{
				Match: effects.And(effects.Creature(), effects.OpponentControls()), Power: -2,
				Label: "Turn the Tide — creatures your opponents control get -2/-0",
			},
		),
	})
}
