package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Piranha Fly — Creature — Fish Insect {1}{U}, 2/1 (Duskmourn: House of
// Horror, #4466):
//
//	"Flying
//	 This creature enters tapped."
//
// "Enters tapped" is the self-replacement every tapland uses
// (SelfEntersTapped, CR 614.1c), so the creature is never untapped on
// the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "58d87cfe-7402-43c3-81ce-13de8cd04005",
		Name:            "Piranha Fly",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements:    []game.ReplacementEffect{SelfEntersTapped()},
	})
}
