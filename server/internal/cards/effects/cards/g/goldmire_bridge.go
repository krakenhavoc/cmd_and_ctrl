package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goldmire Bridge — Artifact Land:
//
//	"This land enters tapped.
//	 Indestructible
//	 {T}: Add {W} or {B}."
//
// Darkmoss Bridge's shape in white and black.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c9b7ea9c-3bcb-4538-aa25-cdb82a52037e",
		Name:            "Goldmire Bridge",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Replacements:    []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities:   []effects.ManaAbility{effects.DualManaAbility("W", "B")},
	})
}
