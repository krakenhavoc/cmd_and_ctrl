package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glistener Seer — Creature — Phyrexian Advisor {U}, 0/3 (Phyrexia: All Will
// Be One):
//
//	"This creature enters with three oil counters on it.
//	 {T}, Remove an oil counter from this creature: Scry 1."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bcbb318d-0ebc-48a2-afd7-69cf5351893d",
		Name:         "Glistener Seer",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{b10EntersWithCounters(oilCounter, 3, "Glistener Seer: enters with three oil counters on it")},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Remove an oil counter from this creature: Scry 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Scry: 1},
			Cost:    Plus(TapCost(), RemoveCountersFromThis(oilCounter, 1)),
			Effect:  Do(Scry{N: 1}),
		}},
	})
}
