package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Henge Guardian — Artifact Creature — Dragon Wurm {5}, 3/4 (Mercadian
// Masques, S82 OLD b681, #5619):
//
//	"{2}: This creature gains trample until end of turn."
//
// The shared self-grant row (Cloudheath Drake's shape) for trample.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "71744428-65ae-42ac-893d-0802fd41e9b4",
		Name:         "Henge Guardian",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.S82OldB510ThisGainsRow("{2}", "trample")},
	})
}
