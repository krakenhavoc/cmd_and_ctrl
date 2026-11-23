package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scion of Ugin — Creature — Dragon Spirit {6}, 4/4 (Dragons of Tarkir,
// #5551):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data (CR 702.9).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7422f43d-0c1e-42cb-864d-1c63f2ec533a",
		Name:            "Scion of Ugin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
