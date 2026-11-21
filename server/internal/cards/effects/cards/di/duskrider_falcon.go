package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Duskrider Falcon — Creature — Bird {1}{W}, 1/1 (Weatherlight, #5629):
//
//	"Flying, protection from black"
//
// Both are engine keywords carried as printed data: flying (CR 702.9b)
// and protection from black (CR 702.16b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d87ab916-5d49-4840-8d5b-805af800a83c",
		Name:            "Duskrider Falcon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "protection from black"},
	})
}
