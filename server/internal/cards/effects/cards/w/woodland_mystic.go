package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Woodland Mystic — Creature — Elf Druid {1}{G}, 1/1 (Core Set 2020, #5518):
//
//	"{T}: Add {G}."
//
// Llanowar Elves' mana ability: a single fixed {G}, so it adds the mana
// at once. Summoning sickness applies (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6996c354-d3fa-40c3-8534-27d2f5b74756",
		Name:         "Woodland Mystic",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
	})
}
