package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Misshapen Fiend — Creature — Horror Mercenary {1}{B}, 1/1 (Mercadian
// Masques, S82 OLD b683, #5619):
//
//	"Flying"
//
// Flying is a printed keyword (CR 702.9a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "10736ea3-6253-44c3-8a1e-3b9a7f38f3cf",
		Name:            "Misshapen Fiend",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
