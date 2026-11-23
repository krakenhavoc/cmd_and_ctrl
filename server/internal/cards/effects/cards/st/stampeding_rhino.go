package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stampeding Rhino — Creature — Rhino {4}{G}, 4/4 (Magic 2010, #5585):
//
//	"Trample (This creature can deal excess combat damage to the player
//	 or planeswalker it's attacking.)"
//
// Trample is an engine keyword carried as printed data (CR 702.19b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8e4dc1fc-e412-4aa3-8793-5c4abc598521",
		Name:            "Stampeding Rhino",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
