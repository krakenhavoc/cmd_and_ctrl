package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sulfurous Mire — Snow Land — Swamp Mountain (Kaldheim, #5505):
//
//	"({T}: Add {B} or {R}.)
//	 This land enters tapped."
//
// The Rakdos member of Kaldheim's snow-dual cycle, written as Ice
// Tunnel is: the snow supertype rides the printed type line, and the
// two colours are declared because two basic land types on a nonbasic
// land produce nothing on their own in this engine.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "77cf536e-246e-4c18-8e04-c904fcad3f40",
		Name:          "Sulfurous Mire",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility("B", "R")},
	})
}
