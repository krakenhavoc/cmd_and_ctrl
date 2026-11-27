package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lightning Hounds — Creature — Dog {2}{R}{R}, 3/2 (Mercadian Masques, S82 OLD
// b684, #5619):
//
//	"First strike"
//
// First strike is a printed keyword (CR 702.7a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b50c2f63-750c-4cfc-8a74-5a7a3d9b46c4",
		Name:            "Lightning Hounds",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
	})
}
