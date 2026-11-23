package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Undersea Invader — Creature — Giant Rogue {4}{U}{U}, 5/6 (Foundations Jumpstart):
//
//	"Flash
//	 This creature enters tapped."
//
// The enters-tapped replacement is the card's own (CR 614.1d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ff12dfee-7140-44bd-a62e-5f8ab98e93c6",
		Name:            "Undersea Invader",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Replacements:    []game.ReplacementEffect{effects.SelfEntersTapped()},
	})
}
