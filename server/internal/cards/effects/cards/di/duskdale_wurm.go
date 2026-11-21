package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Duskdale Wurm — Creature — Wurm {5}{G}{G}, 7/7 (Eventide, #5590):
//
//	"Trample (This creature can deal excess combat damage to the player or
//	 planeswalker it's attacking.)"
//
// The printed keyword (CR 702.19).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "aba4e290-b597-4c93-b3dd-745ca45ad72d",
		Name:            "Duskdale Wurm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
