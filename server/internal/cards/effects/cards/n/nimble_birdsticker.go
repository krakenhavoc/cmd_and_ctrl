package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nimble Birdsticker — Creature — Goblin {2}{R}, 2/3 (Core Set 2020, #5518):
//
//	"Reach (This creature can block creatures with flying.)"
//
// A printed keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2f047fcb-c198-44ac-aa1e-31447f6c6aba",
		Name:            "Nimble Birdsticker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
	})
}
