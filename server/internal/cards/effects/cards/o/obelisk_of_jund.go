package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Obelisk of Jund — Artifact {3} (Shards of Alara, #5588):
//
//	"{T}: Add {B}, {R}, or {G}."
//
// A three-colour mana rock on the tri-lands' b03TriMana row: one tap
// mana ability (CR 605.1a) that adds one mana of a colour its controller
// picks.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "eb0be178-55f9-4743-8128-34bc86f134ec",
		Name:          "Obelisk of Jund",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{effects.B03TriMana("B", "R", "G")},
	})
}
