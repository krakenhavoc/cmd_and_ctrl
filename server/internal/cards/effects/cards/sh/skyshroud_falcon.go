package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skyshroud Falcon — Creature — Bird {1}{W}, 1/1 (Stronghold, S82 OLD b749, #5627):
//
//	"Flying, vigilance"
//
// Both are engine keywords carried as printed data (CR 702.9b, 702.20b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "877ab494-da90-4c26-a184-c6ba64ec30b5",
		Name:            "Skyshroud Falcon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
	})
}
