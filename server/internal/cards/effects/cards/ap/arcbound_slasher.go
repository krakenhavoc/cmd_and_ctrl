package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arcbound Slasher — Artifact Creature — Cat {4}{R}, 0/0 (Modern Horizons
// 2, #5501):
//
//	"Modular 4 (...)
//	 Riot (This creature enters with your choice of an additional +1/+1
//	 counter or haste.)"
//
// Both are engine keywords: modular (CR 702.43a) and riot (CR 702.136a),
// each an entry replacement, applied together.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bf86376d-474d-4658-8cf3-d0e2a01cdfff",
		Name:            "Arcbound Slasher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"modular 4", "riot"},
	})
}
