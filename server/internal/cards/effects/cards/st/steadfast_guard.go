package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Steadfast Guard — Creature — Human Rebel {W}{W}, 2/2 (Mercadian Masques, S82
// OLD b683, #5619):
//
//	"Vigilance"
//
// Vigilance is a printed keyword: attacking does not tap it (CR 702.20b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0c3b3c04-017d-4741-90c8-4b6c0bd46e49",
		Name:            "Steadfast Guard",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
}
