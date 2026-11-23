package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Serra Sphinx — Creature — Sphinx {3}{U}{U}, 4/4 (Planar Chaos, #5595):
//
//	"Flying, vigilance"
//
// Both are engine keywords carried as printed data: flying (CR 702.9b)
// and vigilance (CR 702.20b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9b344428-ec35-45b9-9714-9ca58a338f5c",
		Name:            "Serra Sphinx",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
	})
}
