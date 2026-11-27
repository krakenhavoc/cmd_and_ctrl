package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Perilous Shadow — Creature — Insect Shade {2}{B}{B}, 0/4 (Return to
// Ravnica, S82 OLD):
//
//	"{1}{B}: This creature gets +2/+2 until end of turn."
//
// The Shade's pump is the shared self-pump row; each activation stacks.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3ba3405d-8c9a-4a80-811d-529cd3cb657f",
		Name:         "Perilous Shadow",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.ThisGetsRow("{1}{B}", 2, 2, "Perilous Shadow")},
	})
}
