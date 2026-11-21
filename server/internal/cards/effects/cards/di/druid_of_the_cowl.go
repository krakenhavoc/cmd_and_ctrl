package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Druid of the Cowl — Creature — Elf Druid {1}{G}, 1/3 (Foundations,
// #4465):
//
//	"{T}: Add {G}."
//
// Llanowar Elves' mana ability on a sturdier body. Summoning sickness
// applies to it as to any creature's tap ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b1793c3b-25d6-4fae-a99d-cfdd2210ca67",
		Name:         "Druid of the Cowl",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
	})
}
