package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scragnoth — Creature — Beast {4}{G}, 3/4 (Tempest, S82 OLD batch 767, #5628):
//
//	"This spell can't be countered.
//	 Protection from blue"
//
// "Can't be countered" is the spec's CantBeCountered, working on the stack (CR
// 101.2, CR 113.6g); protection from blue is a printed keyword (CR 702.16).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9e66f2d6-cc51-4747-914d-2c04e439da2b",
		Name:            "Scragnoth",
		Completeness:    effects.CompletenessFull,
		CantBeCountered: true,
		PrintedKeywords: []string{"protection from blue"},
	})
}
