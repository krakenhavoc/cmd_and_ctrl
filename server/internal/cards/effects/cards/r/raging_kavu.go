package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Raging Kavu — Creature — Kavu {1}{R}{G}, 3/1 (Invasion, S82 #5616):
//
//	"Flash
//	 Haste"
//
// Flash and haste are printed keywords (CR 702.8, 702.10).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d126bb08-be60-4046-91f1-65672ef42c63",
		Name:            "Raging Kavu",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "haste"},
	})
}
