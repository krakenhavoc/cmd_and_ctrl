package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seraph of Dawn — {2}{W}{W} Creature — Angel 2/4 (Avacyn Restored, #5570):
//
//	"Flying, lifelink"
//
// Two printed keywords (CR 702.9, CR 702.15).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "38c588ae-7254-4fae-aa9a-03e2a5524492",
		Name:            "Seraph of Dawn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
	})
}
