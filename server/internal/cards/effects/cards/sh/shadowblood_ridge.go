package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shadowblood Ridge — Land (EDHREC top 1,000, #3654):
//
//	"{1}, {T}: Add {B}{R}."
//
// A filter land: no colourless option, one ability, {1} to filter into
// a fixed two colours. The Signet cycle's mana component
// (ManaAbilityCost.Mana) on a land, as Darkwater Catacombs is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "15687ee3-3cdb-4a8f-a726-46b73bceb792",
		Name:         "Shadowblood Ridge",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true, Mana: "{1}"},
			Produced: "{B}{R}",
			Label:    "{1}, {T}: Add {B}{R}",
		}},
	})
}
