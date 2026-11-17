package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tidepool Turtle — Creature — Turtle {3}{U}, 2/5 (Dominaria United, S82):
//
//	"{2}{U}: Scry 1."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ac2668a3-7a26-47b1-9ad4-56e23643de80",
		Name:         "Tidepool Turtle",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{U}: Scry 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Scry: 1},
			Cost:    ManaCost("{2}{U}"),
			Effect:  Do(Scry{N: 1}),
		}},
	})
}
