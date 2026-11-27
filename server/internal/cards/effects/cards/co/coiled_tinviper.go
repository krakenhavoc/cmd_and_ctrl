package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Coiled Tinviper — {3} Artifact Creature — Snake, 2/1 (Tempest, S82 OLD batch
// 770, #5628):
//
//	"First strike"
//
// First strike is a printed keyword (CR 702.7).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b0bdc033-29f0-4dcb-87ff-ccd56126f7e5",
		Name:            "Coiled Tinviper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
	})
}
