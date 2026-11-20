package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boreal Druid — Snow Creature — Elf Druid {G}, 1/1 (EDHREC rank
// 2456):
//
//	"{T}: Add {C}."
//
// Llanowar Elves for a colourless mana — the Elf that pays for
// Eldrazi. The snow supertype is printed data on the type line and
// needs no catalog help; the {C} is a fixed slot, so activating it
// adds mana at once with no colour pick. Summoning sickness applies
// through Game.ActivateManaAbility (CR 302.6), as for every creature
// mana ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2fcc69ff-8ab5-4e14-afe3-db892049a872",
		Name:         "Boreal Druid",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
	})
}
