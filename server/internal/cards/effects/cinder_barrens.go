package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cinder Barrens — Land (Oath of the Gatewatch, #5545):
//
//	"This land enters tapped.
//	 {T}: Add {B} or {R}."
//
// The Guildgates' shape: it enters tapped (a replacement, CR 614.1c) and
// taps for either colour.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "f155a05f-9f5e-4875-a407-103b85ce30ee",
		Name:          "Cinder Barrens",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility("B", "R")},
	})
}
