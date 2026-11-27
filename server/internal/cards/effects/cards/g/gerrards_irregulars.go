package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gerrard's Irregulars — Creature — Human Soldier {4}{R}, 4/2 (Mercadian
// Masques, S82 OLD b683, #5619):
//
//	"Trample, haste"
//
// Both are printed keywords (CR 702.19a, 702.10a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "595cbe23-4969-4fc2-8386-e5f3e6adfd5f",
		Name:            "Gerrard's Irregulars",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
	})
}
