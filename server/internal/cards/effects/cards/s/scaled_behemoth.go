package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scaled Behemoth — Creature — Crocodile {4}{G}{G}, 6/7
// (Amonkhet, #5535):
//
//	"Hexproof (This creature can't be the target of spells or abilities your opponents control.)"
//
// A printed keyword the engine runs off the card; the Spec declares it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "abb39d0e-bf11-4982-b06a-ca7be16bce2c",
		Name:            "Scaled Behemoth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
	})
}
