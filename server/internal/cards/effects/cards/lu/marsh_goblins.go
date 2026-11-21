package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Marsh Goblins — Creature — Goblin {B}{R}, 1/1
// (The Dark, #5638): "Swampwalk". A printed keyword and nothing else.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a2c5a504-5c8f-4c1b-a8b1-4e045aee4673",
		Name:            "Marsh Goblins",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
	})
}
