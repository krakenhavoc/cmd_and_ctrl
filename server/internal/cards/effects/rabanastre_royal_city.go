package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rabanastre, Royal City — Land — Town (Final Fantasy, #4462):
//
//	"This land enters tapped.
//	 {T}: Add {R} or {W}."
//
// The red-white tapped Town, Treno's shape: the enters-tapped self
// replacement (CR 614.1d) and the declared two-colour mana ability. The
// Town subtype rides the printed type line.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "bd39e333-b0b8-4a85-8645-d77f4a3941db",
		Name:          "Rabanastre, Royal City",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility("R", "W")},
	})
}
