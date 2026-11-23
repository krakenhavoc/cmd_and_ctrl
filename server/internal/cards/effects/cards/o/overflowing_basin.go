package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Overflowing Basin — Land:
//
//	"{1}, {T}: Add {G}{U}."
//
// Ferrous Lake's Simic twin: one mana ability costing {1} plus the tap
// (ManaAbilityCost.Mana, paid from the pool) and producing the fixed
// pair {G}{U}. With nothing in the pool to pay the {1} it makes no
// mana, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5ac8e01c-b0a7-4855-a122-1cd26b07c4a5",
		Name:         "Overflowing Basin",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true, Mana: "{1}"},
			Produced: "{G}{U}",
			Label:    "{1}, {T}: Add {G}{U}",
		}},
	})
}
