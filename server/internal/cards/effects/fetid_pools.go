package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fetid Pools — Land — Island Swamp:
//
//	"({T}: Add {U} or {B}.)
//	 This land enters tapped.
//	 Cycling {2}"
//
// The Amonkhet cycling dual. The reminder-text mana ability is
// declared, not derived (the synthetic basic-land ability only fires
// for the BASIC supertype), exactly as the Triomes do; cycling is the
// shared hand ability from cycling.go, and the tapped entry is the
// self-replacement rather than an ETB tap.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "32b03b48-06da-4a74-a7ac-e5ae39a4f428",
		Name:          "Fetid Pools",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility("U", "B")},
		Activated:     []ActivatedAbility{Cycling("{2}")},
	})
}
