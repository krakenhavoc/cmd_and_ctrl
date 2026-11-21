package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spiked Baloth — Creature — Beast {3}{G}, 4/2 (Magic 2013, #5568):
//
//	"Trample (This creature can deal excess combat damage to the player or
//	 planeswalker it's attacking.)"
//
// A keyword body: trample (CR 702.19).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e873b9c2-1346-4f36-aa48-da58112792db",
		Name:            "Spiked Baloth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
