package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mantis Rider — Creature — Human Monk {U}{R}{W}, 3/3 (Khans of Tarkir,
// #5554):
//
//	"Flying, vigilance, haste"
//
// Three engine keywords carried as printed data: flying (CR 702.9b),
// vigilance (CR 702.20b) and haste (CR 702.10b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5c1a0a7a-dea4-4b8e-b9c0-8c19942a01ef",
		Name:            "Mantis Rider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance", "haste"},
	})
}
