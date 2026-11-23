package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gyre Engineer — Creature — Vedalken Wizard {1}{G}{U}, 1/1 (Ravnica
// Allegiance, #5521):
//
//	"{T}: Add {G}{U}."
//
// A mana ability (CR 605.1a) that adds two mana at once, Arixmethes's
// {G}{U}. Summoning sickness gates the {T} as for any creature (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "27eef5c1-259a-4d37-93af-7451f50905af",
		Name:         "Gyre Engineer",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}{U}",
			Label:    "Add {G}{U}",
		}},
	})
}
