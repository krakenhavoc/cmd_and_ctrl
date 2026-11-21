package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Terror of Serpent's Pass — Legendary Creature — Serpent {5}{U}{U}, 8/8 (Avatar: The Last
// Airbender Eternal, #5075):
//
//	"Hexproof"
//
// Hexproof is an engine keyword, carried as printed data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "771bf51d-cb32-4240-936c-f9d1e0c90cf0",
		Name:            "The Terror of Serpent's Pass",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
	})
}
