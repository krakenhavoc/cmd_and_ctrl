package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Transdimensional Bovine — {2}{G} Creature — Ox Avatar 0/4 (Teenage Mutant
// Ninja Turtles, #4457):
//
//	"Flying
//	 {T}: Add two mana of any one color."
//
// Goldspan Dragon's Treasure ability on a creature: one colour picked, two
// of it (OneColorOfAmount). As a creature's {T} ability it waits out
// summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "eaa6b294-ac66-4fe2-b694-97b0a04f7be1",
		Name:            "Transdimensional Bovine",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: effects.OneColorOfAmount(2),
			Label:    "Add two mana of any one color",
		}},
	})
}
