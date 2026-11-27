package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Satyr Hedonist — Creature — Satyr {1}{G}, 2/1 (Theros, S82 old sets):
//
//	"{R}, Sacrifice this creature: Add {R}{R}{R}."
//
// Vessel of Volatility's mana ability with a mana-and-sacrifice cost: it
// never uses the stack and the three red mana are in the pool at once
// (CR 605.3b). No {T} in the cost, so summoning sickness does not apply.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a75bf714-ed3e-4403-a009-3a81f430bc64",
		Name:         "Satyr Hedonist",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Mana: "{R}", Sacrifice: true},
			Produced: "{R}{R}{R}",
			Label:    "{R}, Sacrifice this creature: Add {R}{R}{R}",
		}},
	})
}
