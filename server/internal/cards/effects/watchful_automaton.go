package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Watchful Automaton — Artifact Creature — Construct {3}, 2/2 (Aether Revolt,
// #5536):
//
//	"{2}{U}: Scry 1. (Look at the top card of your library. You may put
//	 that card on the bottom.)"
//
// Tidepool Turtle's scry row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6f5e6799-cb0f-4403-9c2a-69abeac08b79",
		Name:         "Watchful Automaton",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{U}: Scry 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Scry: 1},
			Cost:    ManaCost("{2}{U}"),
			Effect:  Do(Scry{N: 1}),
		}},
	})
}
