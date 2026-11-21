package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spotted Griffin — Creature — Griffin {3}{W}, 2/3 (Portal, #5630):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data: it can be blocked
// only by creatures with flying or reach (CR 702.9b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4916773d-5ccb-48ff-8aa3-09771ae88e81",
		Name:            "Spotted Griffin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
