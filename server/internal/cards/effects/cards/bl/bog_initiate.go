package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bog Initiate — Creature — Human Wizard {1}{B}, 1/1 (Invasion, #5616):
//
//	"{1}: Add {B}."
//
// A mana ability (CR 605.1a) whose whole cost is {1}; no {T}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "23f93411-f83c-4ed2-abed-99cf905f7d7f",
		Name:         "Bog Initiate",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Mana: "{1}"},
			Produced: "{B}",
			Label:    "{1}: Add {B}",
		}},
	})
}
