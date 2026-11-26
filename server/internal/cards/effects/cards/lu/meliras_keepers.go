package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Melira's Keepers — Creature — Human Warrior {4}{G}, 4/4 (Mirrodin Besieged,
// S82 OLD b467, #5576):
//
//	"This creature can't have counters put on it."
//
// A layer 6 static on itself that sets the characteristic the engine reads
// everywhere a counter would be placed, paid or chosen (CR 101.2): +1/+1 and
// -1/-1 counters, proliferate and counter costs all leave it out. It is a
// printed ability, so it is lost when the Keepers loses its abilities.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6f2f4c31-74dc-4a52-96ad-94c3d51398d1",
		Name:         "Melira's Keepers",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature can't have counters put on it.",
			Layer:     game.Layer6Ability,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.CantHaveCounters = true
			},
		}},
	})
}
