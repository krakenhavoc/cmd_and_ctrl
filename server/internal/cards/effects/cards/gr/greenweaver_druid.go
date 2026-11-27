package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Greenweaver Druid — Creature — Elf Druid {2}{G}, 1/1 (Zendikar, #5583):
//
//	"{T}: Add {G}{G}."
//
// A {T} mana ability (CR 605.1a) that produces two green mana as one
// ability, so a creature that has not been under its controller's control
// since the start of the turn can't use it (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5e35a7c9-1e6e-41a7-9f59-5ea1bc53bc65",
		Name:         "Greenweaver Druid",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}{G}",
			Label:    "Add {G}{G}",
		}},
	})
}
