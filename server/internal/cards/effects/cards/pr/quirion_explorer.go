package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quirion Explorer — Creature — Elf Druid Scout {1}{G}, 1/1 (Commander 2016,
// #5537):
//
//	"{T}: Add one mana of any color that a land an opponent controls could
//	 produce."
//
// Exotic Orchard's derived mana ability on a creature: the colour is read from
// the opposing lands' mana abilities as the ability is activated (CR 106.7),
// colourless is not a colour, and no opposing land that could produce a colour
// means no mana. Summoning sickness gates the {T} as for any creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "766c8a68-9c51-44ea-903c-1f5cb30941d1",
		Name:         "Quirion Explorer",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:                    effects.ManaAbilityCost{Tap: true},
			DerivedMatch:            effects.DerivedFromOpponentLands(),
			DerivedColorsOnly:       true,
			DerivesFromOtherSources: true,
			Label:                   "Add one mana of any color an opponent's land could produce",
		}},
	})
}
