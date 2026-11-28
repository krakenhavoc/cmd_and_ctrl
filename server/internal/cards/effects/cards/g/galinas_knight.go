package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Galina's Knight — Creature — Merfolk Knight {W}{U}, 2/2 (Invasion, S82 #5616):
//
//	"Protection from red"
//
// Protection from red is a printed keyword (CR 702.16).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9a484eb3-ae4a-465b-9dbc-f7a42a036e07",
		Name:            "Galina's Knight",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from red"},
	})
}
