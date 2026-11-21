package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moon Heron — Creature — Spirit Bird {3}{U}, 3/2 (Innistrad, #5572):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data: only a creature
// with flying or reach can block it (CR 702.9b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "57dfe4aa-1e9a-4891-851b-201edd1fc79c",
		Name:            "Moon Heron",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
