package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Verdant Automaton — Artifact Creature — Construct {2}, 1/2 (Aether Revolt,
// #5536):
//
//	"{3}{G}: Put a +1/+1 counter on this creature."
//
// Jungle Delver's counter row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d3af1c69-42ae-4d29-a9f1-72f862d26303",
		Name:         "Verdant Automaton",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}{G}: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{3}{G}"),
			Effect:  Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}),
		}},
	})
}
