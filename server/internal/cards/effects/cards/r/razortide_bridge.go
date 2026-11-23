package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Razortide Bridge — Artifact Land (Commander deck reprint):
//
//	"This land enters tapped.
//	 Indestructible
//	 {T}: Add {W} or {U}."
//
// Thornglint Bridge's shape with another colour pair.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6cb37ac1-dd11-4a8c-bca5-ef44d828059f",
		Name:            "Razortide Bridge",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Replacements:    []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities:   []effects.ManaAbility{effects.DualManaAbility("W", "U")},
	})
}
