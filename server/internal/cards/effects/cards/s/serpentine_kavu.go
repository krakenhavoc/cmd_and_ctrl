package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Serpentine Kavu (Invasion, S82 OLD batch 647, #5616):
//
//	Creature — Kavu {4}{G}, 4/4:
//	"{R}: This creature gains haste until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf252d71-77e7-438f-b774-28e5ab2f09f6",
		Name:         "Serpentine Kavu",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.S82OldB510ThisGainsRow("{R}", "haste")},
	})
}
