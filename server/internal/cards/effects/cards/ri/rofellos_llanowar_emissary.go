package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rofellos, Llanowar Emissary — Legendary Creature — Elf Druid {G}{G},
// 2/1 (Urza's Destiny, #5621):
//
//	"{T}: Add {G} for each Forest you control."
//
// Magus of the Coffers' count with no mana in the cost: a mana ability
// (CR 605.1a) that adds one {G} per permanent with the Forest subtype
// its controller controls as it resolves, basic or not. On a creature
// the {T} needs it to have been under its controller's control since
// their most recent turn began (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3015882b-4897-4d2c-8e33-6731c85a0d03",
		Name:         "Rofellos, Llanowar Emissary",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:           effects.ManaAbilityCost{Tap: true},
			ProducedAmount: effects.ProducedPerPermanent("G", effects.MatchLandSubtype("Forest")),
			Label:          "Add {G} for each Forest you control",
		}},
	})
}
