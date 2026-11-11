package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "6cb37ac1-dd11-4a8c-bca5-ef44d828059f",
		Name:            "Razortide Bridge",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Replacements:    []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities:   []ManaAbility{dualManaAbility("W", "U")},
	})
}
