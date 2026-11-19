package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mox Poison — Artifact {0} (Mystery Booster 2):
//
//	"{T}: Add one mana of any color. You get two poison counters."
//
// The poison is part of the mana ability, so it is a rider that lands
// with the mana (CR 605.3b), not a trigger and not a cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8d1a45b1-08c4-441d-86d6-f44061230c0b",
		Name:         "Mox Poison",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}: Add one mana of any color. You get two poison counters.",
			Rider: func(g *game.Game, controller, source uuid.UUID) error {
				return g.AddPlayerCounterByForEffect(controller, controller, game.CounterPoison, 2)
			},
		}},
	})
}
