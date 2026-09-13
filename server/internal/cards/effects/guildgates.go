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
// Two of the ten are in the roadmap's batch 14 (#307), two more in
// batch 16 (#309) and one in batch 17 (#310); the other five belong
// in this table when their batches reach them — a card
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
		// Roadmap batch 16 (#309).
		{"bf75a3d1-f184-4b48-a913-21caee1db084", "Izzet Guildgate", "U", "R"},
		{"52d14717-0cbc-4d7e-b546-54ea91580338", "Dimir Guildgate", "U", "B"},
		// Roadmap batch 17 (#310).
		{"73c423b7-cab8-4e69-8070-9edbf96a6c2c", "Boros Guildgate", "R", "W"},
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
