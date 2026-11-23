package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wasp Lancer — Creature — Faerie Soldier {U/B}{U/B}{U/B}, 3/2
// (Shadowmoor, #5591):
//
//	"Flying"
//
// Flying is a printed keyword (CR 702.9b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "12a9a9a4-6b8c-48a1-833f-457a7ece8e7e",
		Name:            "Wasp Lancer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
