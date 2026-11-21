package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arboretum Elemental — Creature — Elemental {7}{G}{G}, 7/5 (Guilds of Ravnica, #5525):
//
//	"Convoke
//	 Hexproof"
//
// Convoke is the shared tap cost (CR 702.51a) and hexproof is printed
// data (CR 702.11b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1708d60d-30af-4184-bc95-2f7a7aed4a6e",
		Name:            "Arboretum Elemental",
		Completeness:    effects.CompletenessFull,
		TapCost:         effects.Convoke(),
		PrintedKeywords: []string{"hexproof"},
	})
}
