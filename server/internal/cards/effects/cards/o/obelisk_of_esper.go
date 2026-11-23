package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Obelisk of Esper — Artifact {3} (Shards of Alara, #5588):
//
//	"{T}: Add {W}, {U}, or {B}."
//
// A three-colour mana rock on the tri-lands' b03TriMana row: one tap
// mana ability (CR 605.1a) that adds one mana of a colour its controller
// picks.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "9da963f6-734b-4d3c-a3be-dd5cca2c7b19",
		Name:          "Obelisk of Esper",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{effects.B03TriMana("W", "U", "B")},
	})
}
