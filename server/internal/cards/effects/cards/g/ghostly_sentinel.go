package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghostly Sentinel — Creature — Kor Spirit {4}{W}, 3/3 (Battle for
// Zendikar, #5547):
//
//	"Flying, vigilance"
//
// Flying and vigilance are engine keywords carried as printed data: it
// can be blocked only by creatures with flying or reach (CR 702.9b),
// and attacking doesn't tap it (CR 702.20b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a16f4b53-c8cf-45b0-89e7-507acd094a64",
		Name:            "Ghostly Sentinel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
	})
}
