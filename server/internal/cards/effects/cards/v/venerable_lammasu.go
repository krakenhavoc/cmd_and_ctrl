package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Venerable Lammasu — Creature — Lammasu {6}{W}, 5/4 (Khans of Tarkir,
// #5554):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data (CR 702.9b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "359c6272-fcb5-4d3f-a4b5-4209e354a044",
		Name:            "Venerable Lammasu",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
