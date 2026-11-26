package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Phyrexian Vatmother — Creature — Phyrexian Horror {2}{B}{B}, 4/5 (Mirrodin
// Besieged, S82 OLD b466, #5576):
//
//	"Infect
//	 At the beginning of your upkeep, you get a poison counter."
//
// The controller puts the counter on themselves (CR 122.1f) every upkeep of
// their own turn, as the trigger resolves. Infect is the engine's damage
// tail (ADR 0056).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e9162478-0b10-422a-a11c-7b39b00f026e",
		Name:            "Phyrexian Vatmother",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"infect"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Phyrexian Vatmother — you get a poison counter",
				func(g *game.Game, item *game.StackItem) error {
					return g.AddPlayerCounterByForEffect(item.Controller, item.Controller, game.CounterPoison, 1)
				}),
		},
	})
}
