package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Looming Shade — Creature — Shade {2}{B}, 1/1 (Magic 2010, #5585):
//
//	"{B}: This creature gets +1/+1 until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3ce8560d-dc68-4551-81c9-3a7b8d8da9a1",
		Name:         "Looming Shade",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.SelfPumpRow("Looming Shade", "{B}", 1, 1)},
	})
}
