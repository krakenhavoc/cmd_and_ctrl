package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hornet Cobra — Creature — Snake {1}{G}{G}, 2/1 (Legends, #5639):
//
//	"First strike"
//
// First strike is an engine keyword: it deals its combat damage in the
// first-strike damage step (CR 702.7b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4958baf3-2f5a-4087-bc14-ffee9887db1d",
		Name:            "Hornet Cobra",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
	})
}
