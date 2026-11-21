package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Weaver of Currents — Creature — Snake Druid {1}{G}{U}, 2/2
// (Amonkhet, #5535):
//
//	"{T}: Add {C}{C}."
//
// A tap mana ability, the shape of Worn Powerstone's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9aabb520-7e97-47b2-975c-11f457d0b870",
		Name:         "Weaver of Currents",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}{C}",
			Label:    "Add {C}{C}",
		}},
	})
}
