package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rustvale Bridge — Artifact Land (Commander deck reprint):
//
//	"This land enters tapped.
//	 Indestructible
//	 {T}: Add {R} or {W}."
//
// Thornglint Bridge's shape with another colour pair.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a3faf70d-c034-4692-9e92-1922029e3852",
		Name:            "Rustvale Bridge",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Replacements:    []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities:   []ManaAbility{dualManaAbility("R", "W")},
	})
}
