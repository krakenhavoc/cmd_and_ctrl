package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Armored Griffin — Creature — Griffin {3}{W}, 2/3 (Portal Second Age,
// #5625):
//
//	"Flying, vigilance"
//
// Flying and vigilance are engine keywords carried as printed data: it
// can be blocked only by creatures with flying or reach (CR 702.9b), and
// attacking doesn't tap it (CR 702.20b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ad3d6003-66d3-486d-aa54-ddce1adb5ff1",
		Name:            "Armored Griffin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
	})
}
