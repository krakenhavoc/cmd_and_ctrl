package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "c9b7ea9c-3bcb-4538-aa25-cdb82a52037e",
		Name:            "Goldmire Bridge",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Replacements:    []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities:   []ManaAbility{dualManaAbility("W", "B")},
	})
}
