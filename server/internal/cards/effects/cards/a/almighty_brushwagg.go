package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Almighty Brushwagg — Creature — Brushwagg {G}, 1/1 (Ikoria: Lair of Behemoths, #5512):
//
//	"Trample
//	 {3}{G}: This creature gets +3/+3 until end of turn."
//
// The firebreathing row (selfPumpRow): no tap, so any number of times.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d8a3a89d-9d27-40ae-a1aa-d55cc6a79b0b",
		Name:            "Almighty Brushwagg",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated:       []effects.ActivatedAbility{effects.SelfPumpRow("Almighty Brushwagg", "{3}{G}", 3, 3)},
	})
}
