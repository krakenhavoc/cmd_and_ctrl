package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Diving Griffin — Creature — Griffin {1}{W}{W}, 2/2 (Prophecy, #5617):
//
//	"Flying, vigilance"
//
// Two printed keywords and nothing else. Flying limits its blockers to
// creatures with flying or reach (CR 702.9b), and vigilance means it
// doesn't tap to attack (CR 702.20b); the engine enforces both.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8785f9b4-eea2-4b45-9ae4-194b9a715702",
		Name:            "Diving Griffin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
	})
}
