package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// guildgates.go — the Ravnica Guildgate cycle:
//
//	"This land enters tapped.
//	 {T}: Add {A} or {B}."
//
// The budget dual with a Gate subtype, which is the whole reason a
// deck plays one over a gain land: Gates count for Maze's End and
// Gatecreeper Vine, and the printed type line is what carries that.
// Two of the ten are in the roadmap's batch 14 (#307); the other
// eight belong in this table when their batches reach them — a card
// that belongs to an existing cycle goes in the cycle's table, never
// in a new file.
//
// Enters-tapped is the self-replacement; the dual is the pipe every
// two-colour land in the catalog uses.
//
// No simplification.
func init() {
	for _, land := range []struct{ oracleID, name, a, b string }{
		{"e8705df9-6439-4930-91b6-229f818559af", "Simic Guildgate", "G", "U"},
		{"fa2da325-6859-45bb-b185-35526b01bcc1", "Golgari Guildgate", "B", "G"},
	} {
		Register(Spec{
			OracleID:      land.oracleID,
			Name:          land.name,
			Completeness:  CompletenessFull,
			Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
			ManaAbilities: []ManaAbility{dualManaAbility(land.a, land.b)},
		})
	}
}
