package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jhovall Queen — Creature — Cat Rebel {4}{W}{W}, 4/7 (Mercadian Masques, S82
// OLD b684, #5619):
//
//	"Vigilance"
//
// Vigilance is a printed keyword (CR 702.20b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e1a48505-acbe-4eb4-8e83-3c76cf7a8829",
		Name:            "Jhovall Queen",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
}
