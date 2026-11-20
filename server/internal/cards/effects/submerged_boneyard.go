package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Submerged Boneyard — Land (Oath of the Gatewatch, #5545):
//
//	"This land enters tapped.
//	 {T}: Add {U} or {B}."
//
// The Guildgates' shape: it enters tapped (a replacement, CR 614.1c) and
// taps for either colour.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "f27d52e6-aab9-4f95-ae46-33d1173bf4fe",
		Name:          "Submerged Boneyard",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility("U", "B")},
	})
}
