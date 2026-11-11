package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "2065cada-4078-41c4-9e06-2460d2a2e8ee",
		Name:            "Darkmoss Bridge",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Replacements:    []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities:   []ManaAbility{dualManaAbility("B", "G")},
	})
}
