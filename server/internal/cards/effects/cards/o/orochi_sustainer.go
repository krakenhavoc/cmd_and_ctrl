package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Orochi Sustainer — Creature — Snake Shaman {1}{G}, 1/2 (Champions of
// Kamigawa, #5603):
//
//	"{T}: Add {G}."
//
// A mana ability (CR 605.1a) with {T}, so summoning sickness stops it
// (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d3039b42-25d5-4fe2-955a-0d910d7395bd",
		Name:         "Orochi Sustainer",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
	})
}
