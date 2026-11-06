package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Treno, Dark City — Land — Town (Final Fantasy, #4462):
//
//	"This land enters tapped.
//	 {T}: Add {U} or {B}."
//
// The blue-black tapped Town, Sharlayan's shape: the enters-tapped self
// replacement (CR 614.1d) and the declared two-colour mana ability. The
// Town subtype rides the printed type line.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "6d5bb714-1bf4-4f75-9ea8-bdbd18b0425b",
		Name:          "Treno, Dark City",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility("U", "B")},
	})
}
