package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sharlayan, Nation of Scholars — Land — Town (Final Fantasy, #4462):
//
//	"This land enters tapped.
//	 {T}: Add {W} or {U}."
//
// The white-blue tapped Town. Enters tapped is the self replacement (CR
// 614.1d); the mana ability is the declared two-colour one, since a Town
// has no basic land type to derive it from. The Town subtype rides the
// printed type line, which is what the set's "Towns you control" counts
// read.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "564bdbdd-8392-4ee1-a132-1a17a67b2110",
		Name:          "Sharlayan, Nation of Scholars",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility("W", "U")},
	})
}
