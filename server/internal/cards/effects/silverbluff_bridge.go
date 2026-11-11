package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "081bfd50-a436-463b-9d2c-5bc8a32b387c",
		Name:            "Silverbluff Bridge",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Replacements:    []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities:   []ManaAbility{dualManaAbility("U", "R")},
	})
}
