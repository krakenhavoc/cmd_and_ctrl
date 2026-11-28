package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thermal Glider — Creature — Human Rebel {2}{W}, 2/1 (Mercadian Masques, S82
// OLD b684, #5619):
//
//	"Flying, protection from red"
//
// Both are printed keywords (CR 702.9a, 702.16).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cbe5c74b-1006-4c63-8b46-e06b61e6fe02",
		Name:            "Thermal Glider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "protection from red"},
	})
}
