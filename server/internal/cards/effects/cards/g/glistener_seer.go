package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glistener Seer — Creature — Phyrexian Advisor {U}, 0/3 (Phyrexia: All Will
// Be One):
//
//	"This creature enters with three oil counters on it.
//	 {T}, Remove an oil counter from this creature: Scry 1."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bcbb318d-0ebc-48a2-afd7-69cf5351893d",
		Name:         "Glistener Seer",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.B10EntersWithCounters(effects.OilCounter, 3, "Glistener Seer: enters with three oil counters on it")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Remove an oil counter from this creature: Scry 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Scry: 1},
			Cost:    effects.Plus(effects.TapCost(), effects.RemoveCountersFromThis(effects.OilCounter, 1)),
			Effect:  effects.Do(effects.Scry{N: 1}),
		}},
	})
}
