package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Darkmoss Bridge — Artifact Land:
//
//	"This land enters tapped.
//	 Indestructible
//	 {T}: Add {B} or {G}."
//
// Thornglint Bridge's shape in black and green.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2065cada-4078-41c4-9e06-2460d2a2e8ee",
		Name:            "Darkmoss Bridge",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Replacements:    []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities:   []effects.ManaAbility{effects.DualManaAbility("B", "G")},
	})
}
