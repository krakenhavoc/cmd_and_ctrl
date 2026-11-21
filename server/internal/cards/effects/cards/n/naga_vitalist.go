package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Naga Vitalist — {1}{G} Creature — Snake Druid 1/2 (Jumpstart 2022, #5948):
//
//	"{T}: Add one mana of any type that a land you control could produce."
//
// Reflecting Pool's derived mana: every land you control is asked what
// it could produce, so a Vitalist with no lands makes nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9cea8783-c7d7-49f9-b23d-324309efc140",
		Name:         "Naga Vitalist",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:                    effects.ManaAbilityCost{Tap: true},
			DerivedMatch:            effects.DerivedFromOwnLands(),
			DerivedColorsOnly:       false,
			DerivesFromOtherSources: true,
			Label:                   "Add one mana of any type that a land you control could produce",
		}},
	})
}
