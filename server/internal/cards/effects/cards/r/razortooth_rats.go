package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Razortooth Rats — Creature — Rat {2}{B}, 2/1 (Weatherlight, #5629):
//
//	"Fear (This creature can't be blocked except by artifact creatures
//	 and/or black creatures.)"
//
// Fear is an engine keyword carried as printed data (CR 702.36b), as
// on Shriekmaw.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "72e9a207-dad9-4c51-8040-e90a36bd35e6",
		Name:            "Razortooth Rats",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"fear"},
	})
}
