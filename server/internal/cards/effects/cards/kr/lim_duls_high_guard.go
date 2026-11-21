package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lim-Dûl's High Guard — Creature — Skeleton {1}{B}{B}, 2/1 (Alliances,
// #5634):
//
//	"First strike
//	 {1}{B}: Regenerate this creature."
//
// First strike is an engine keyword carried as printed data (CR 702.7b).
// The ability makes a regeneration shield on itself (CR 701.19a), built by
// regenerateThisRow (regenerate_this.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b0873be6-0b77-4ca4-8c45-a1207b184771",
		Name:            "Lim-Dûl's High Guard",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Activated:       []effects.ActivatedAbility{effects.RegenerateThisRow("{1}{B}")},
	})
}
