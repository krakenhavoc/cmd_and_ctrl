package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silverbluff Bridge — Artifact Land (Commander deck reprint):
//
//	"This land enters tapped.
//	 Indestructible
//	 {T}: Add {U} or {R}."
//
// Thornglint Bridge's shape with another colour pair.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "081bfd50-a436-463b-9d2c-5bc8a32b387c",
		Name:            "Silverbluff Bridge",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Replacements:    []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities:   []effects.ManaAbility{effects.DualManaAbility("U", "R")},
	})
}
