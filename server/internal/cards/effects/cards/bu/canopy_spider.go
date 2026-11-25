package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Canopy Spider — Creature — Spider {1}{G}, 1/3 (Core Set 2020, #5518):
//
//	"Reach (This creature can block creatures with flying.)"
//
// A printed keyword and nothing else.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "37f3733e-cc4e-4d84-b29b-d474f6e254a2",
		Name:            "Canopy Spider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
	})
}
