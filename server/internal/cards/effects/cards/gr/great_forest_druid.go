package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Great Forest Druid — Creature — Treefolk Druid {1}{G}, 0/4 (Lorwyn
// Eclipsed, #4458):
//
//	"{T}: Add one mana of any color."
//
// A mana ability (CR 605.1a): no target, no stack, and it adds one mana of
// a color its controller names as it is activated. Unrestricted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5cb6ebf7-6a75-4464-9e39-e171fcb52de3",
		Name:         "Great Forest Druid",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}: Add one mana of any color.",
		}},
	})
}
