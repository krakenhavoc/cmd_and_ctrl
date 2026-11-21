package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Odious Trow — Creature — Troll {B/G}, 1/1 (Eventide, #5590):
//
//	"{1}{B/G}: Regenerate this creature."
//
// The shared "<cost>: Regenerate this creature." row (regenerateThisRow).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0e9864e2-1bb4-4527-b2af-62433a90b669",
		Name:         "Odious Trow",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.RegenerateThisRow("{1}{B/G}")},
	})
}
