package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tundra Wolves — Creature — Wolf {W}, 1/1 (Legends, #5639):
//
//	"First strike (This creature deals combat damage before creatures without first strike.)"
//
// First strike is an engine keyword: it deals its combat damage in the
// first-strike damage step (CR 702.7b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2155864e-5788-4e72-a800-e2cf25cf59a7",
		Name:            "Tundra Wolves",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
	})
}
