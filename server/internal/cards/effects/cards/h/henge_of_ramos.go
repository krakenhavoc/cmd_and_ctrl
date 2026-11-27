package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Henge of Ramos — Land (Mercadian Masques, #5619):
//
//	"{T}: Add {C}.
//	 {2}, {T}: Add one mana of any color."
//
// Two plain mana abilities, the second paying {2} and tapping (CR 605.1a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "829474df-6413-4323-aef6-f878cb0e797c",
		Name:         "Henge of Ramos",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{Cost: effects.ManaAbilityCost{Tap: true}, Produced: "{C}", Label: "Add {C}"},
			{Cost: effects.ManaAbilityCost{Mana: "{2}", Tap: true}, Produced: "{W|U|B|R|G}", Label: "{2}, {T}: Add one mana of any color"},
		},
	})
}
