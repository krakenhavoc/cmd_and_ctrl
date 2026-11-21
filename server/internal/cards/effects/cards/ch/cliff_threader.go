package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cliff Threader — Creature — Kor Scout {1}{W}, 2/1 (Zendikar, #5583):
//
//	"Mountainwalk (This creature can't be blocked as long as defending
//	 player controls a Mountain.)"
//
// Mountainwalk is landwalk, an engine keyword (CR 702.14c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "27882623-b5ad-49f2-b5ab-f0bdb93268a1",
		Name:            "Cliff Threader",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"mountainwalk"},
	})
}
