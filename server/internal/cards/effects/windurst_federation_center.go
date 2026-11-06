package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Windurst, Federation Center — Land — Town (Final Fantasy, #4462):
//
//	"This land enters tapped.
//	 {T}: Add {G} or {W}."
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
		OracleID:      "e198126a-f280-47f7-8bd0-3dc9f5ff05a0",
		Name:          "Windurst, Federation Center",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility("G", "W")},
	})
}
