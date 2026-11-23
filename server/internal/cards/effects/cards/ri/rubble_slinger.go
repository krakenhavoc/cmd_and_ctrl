package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rubble Slinger — Creature — Human Warrior {2}{R/G}, 2/3 (Ravnica
// Allegiance, #5521):
//
//	"Reach"
//
// Reach is an engine keyword carried as printed data (CR 702.17b): it can
// block a creature with flying.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4de340bf-eab6-4066-896d-7084a9c13c2d",
		Name:            "Rubble Slinger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
	})
}
