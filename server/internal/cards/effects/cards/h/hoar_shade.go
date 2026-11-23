package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hoar Shade — Creature — Shade {3}{B}, 1/2 (Ice Age, #5636):
//
//	"{B}: This creature gets +1/+1 until end of turn."
//
// The shade pump, the shared selfPumpRow at +1/+1 for {B}: no tap, so
// any number of activations, summoning sickness or not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "23c74c42-11bd-4fba-b1cb-93b96f5a11fb",
		Name:         "Hoar Shade",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.SelfPumpRow("Hoar Shade", "{B}", 1, 1)},
	})
}
