package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nightwind Glider — Creature — Human Rebel {2}{W}, 2/1 (Mercadian Masques,
// S82 OLD b684, #5619):
//
//	"Flying, protection from black"
//
// Both are printed keywords: flying (CR 702.9a) and protection from black
// (CR 702.16), which stops damage, enchanting and equipping, blocking by and
// targeting by black sources.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9a642db9-c337-481c-b260-11e01f64e68e",
		Name:            "Nightwind Glider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "protection from black"},
	})
}
