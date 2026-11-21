package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Volcanic Dragon — Creature — Dragon {4}{R}{R}, 4/4 (Mirage, #5633):
//
//	"Flying, haste"
//
// Two printed keywords: flying for blocks (CR 702.9b), haste so it can
// attack the turn it arrives (CR 702.10b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "994db177-03f5-43dd-bf7b-2994e8d430d3",
		Name:            "Volcanic Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
	})
}
