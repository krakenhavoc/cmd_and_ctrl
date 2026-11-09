package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Insomnia, Crown City — Land — Town (Final Fantasy, #4462):
//
//	"This land enters tapped.
//	 {T}: Add {W} or {B}."
//
// The white-black tapped Town, Treno's shape: the enters-tapped self
// replacement (CR 614.1d) and the declared two-colour mana ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "b7140f72-91b6-4be4-8d71-c019e64c2351",
		Name:          "Insomnia, Crown City",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility("W", "B")},
	})
}
