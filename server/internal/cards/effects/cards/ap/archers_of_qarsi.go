package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Archers of Qarsi — Creature — Snake Archer {3}{G}, 5/2 (Fate Reforged,
// #5552):
//
//	"Defender
//	 Reach (This creature can block creatures with flying.)"
//
// Defender (CR 702.3b) and reach (CR 702.17b) are engine keywords carried as
// printed data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ad6bdd24-5a32-4046-83ee-2fe51bf8bb7a",
		Name:            "Archers of Qarsi",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender", "reach"},
	})
}
