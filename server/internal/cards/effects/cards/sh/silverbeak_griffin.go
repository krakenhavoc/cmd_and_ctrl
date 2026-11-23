package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silverbeak Griffin — Creature — Griffin {W}{W}, 2/2 (Core Set 2019, #5527):
//
//	"Flying (This creature can't be blocked except by creatures with flying or reach.)"
//
// The whole text is printed keywords; the engine reads them off
// PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e9f410c1-d93d-4291-95fc-513832fb36e0",
		Name:            "Silverbeak Griffin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
