package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heartwood Treefolk — Creature — Treefolk {2}{G}{G}, 3/4 (Tempest, S82 OLD
// batch 767, #5628):
//
//	"Forestwalk (This creature can't be blocked as long as defending player
//	 controls a Forest.)"
//
// A printed landwalk keyword (CR 702.14): the block legality check looks for a
// Forest among the defending player's lands as the blockers are declared.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "27c21086-8d19-4379-af37-f251b2ec2502",
		Name:            "Heartwood Treefolk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"forestwalk"},
	})
}
