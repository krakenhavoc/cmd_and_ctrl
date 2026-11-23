package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Somberwald Dryad — Creature — Dryad {1}{G}, 2/2 (Dark Ascension, #5571):
//
//	"Forestwalk (This creature can't be blocked as long as defending
//	 player controls a Forest.)"
//
// Forestwalk is the landwalk keyword the block gate reads (CR 702.14), as
// on Zodiac Tiger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b3e9ae89-3507-4429-ab37-7175f4db4710",
		Name:            "Somberwald Dryad",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"forestwalk"},
	})
}
