package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vector, Imperial Capital — Land — Town (Final Fantasy, #4462):
//
//	"This land enters tapped.
//	 {T}: Add {B} or {R}."
//
// Baron, Airship Kingdom's shape: the enters-tapped self-replacement
// (CR 614.1c, so nobody gets a window to answer it) and the two-colour
// mana ability every dual in the catalog uses. Registered because the
// engine's synthetic mana ability only covers basic lands. Town is a
// land subtype with no rules of its own; the type line carries it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "8dcba63c-4701-4dc9-81f5-ca8e8933a3ba",
		Name:          "Vector, Imperial Capital",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility("B", "R")},
	})
}
