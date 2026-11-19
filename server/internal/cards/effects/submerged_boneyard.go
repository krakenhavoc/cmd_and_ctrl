package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Submerged Boneyard — Land:
//
//	"This land enters tapped.
//	 {T}: Add {U} or {B}."
//
// A plain tapped Dimir dual: the tapped entry is the CR 614 self
// replacement, the mana ability is the two-colour tap.
func init() {
	Register(Spec{
		OracleID:     "f27d52e6-aab9-4f95-ae46-33d1173bf4fe",
		Name:         "Submerged Boneyard",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U|B}",
			Label:    "Add {U} or {B}",
		}},
	})
}
