package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ur-Golem's Eye — Artifact {4} (Darksteel, #5605):
//
//	"{T}: Add {C}{C}."
//
// Sol Ring's mana ability (CR 605.1a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fb34fc00-e60d-41fc-9393-ca4248ec0a1c",
		Name:         "Ur-Golem's Eye",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}{C}",
			Label:    "Add {C}{C}",
		}},
	})
}
