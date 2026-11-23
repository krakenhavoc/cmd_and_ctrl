package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kozilek's Channeler — Creature — Eldrazi {5}, 4/4 (Battle for
// Zendikar, #5547):
//
//	"{T}: Add {C}{C}."
//
// Palladium Myr's mana ability on an Eldrazi body: a real mana ability
// (CR 605.1a), so it doesn't use the stack, and the {T} cost waits out
// summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "33826f91-7680-4a4c-be39-db6c2c0bd465",
		Name:         "Kozilek's Channeler",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}{C}",
			Label:    "Add {C}{C}",
		}},
	})
}
