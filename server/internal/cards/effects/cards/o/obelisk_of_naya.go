package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Obelisk of Naya — Artifact {3} (Shards of Alara, #5588):
//
//	"{T}: Add {R}, {G}, or {W}."
//
// A three-colour mana rock on the tri-lands' b03TriMana row: one tap
// mana ability (CR 605.1a) that adds one mana of a colour its controller
// picks.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "02e81bb8-ba80-4483-9b9f-973dcae0fc9d",
		Name:          "Obelisk of Naya",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{effects.B03TriMana("R", "G", "W")},
	})
}
