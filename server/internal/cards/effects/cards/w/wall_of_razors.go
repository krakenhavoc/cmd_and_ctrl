package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wall of Razors — Creature — Wall {1}{R}, 4/1 (Stronghold, S82 OLD b749, #5627):
//
//	"Defender
//	 First strike"
//
// Both are engine keywords carried as printed data: it can't attack (CR
// 702.3b) and deals combat damage in the first-strike step (CR 702.7b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "19ed5429-5631-4a36-a69c-ab35860160b6",
		Name:            "Wall of Razors",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender", "first strike"},
	})
}
