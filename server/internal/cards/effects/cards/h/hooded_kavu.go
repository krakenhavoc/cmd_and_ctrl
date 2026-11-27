package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hooded Kavu (Invasion, S82 OLD batch 645, #5616):
//
//	Creature — Kavu {2}{R}, 2/2:
//	"{B}: This creature gains fear until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "52934f16-eb9b-4d0a-b372-155461660f6e",
		Name:         "Hooded Kavu",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.S82OldB510ThisGainsRow("{B}", "fear")},
	})
}
