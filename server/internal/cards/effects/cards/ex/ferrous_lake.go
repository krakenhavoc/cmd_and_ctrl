package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ferrous Lake — Land:
//
//	"{1}, {T}: Add {U}{R}."
//
// The Izzet filter land with no colourless ability of
// its own: a single mana ability whose cost is {1} plus the tap
// (ManaAbilityCost.Mana, the component Mystic Gate's filter ability
// uses), paid from the pool, producing the fixed pair {U}{R}. With
// nothing in the pool to pay the {1} it makes no mana, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "62c15af0-40e1-407d-b056-7a3d909e3fdb",
		Name:         "Ferrous Lake",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true, Mana: "{1}"},
			Produced: "{U}{R}",
			Label:    "{1}, {T}: Add {U}{R}",
		}},
	})
}
