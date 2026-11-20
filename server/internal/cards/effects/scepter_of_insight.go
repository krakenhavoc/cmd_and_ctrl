package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scepter of Insight — Artifact {1}{U}{U} (Conflux, #5587):
//
//	"{3}{U}, {T}: Draw a card."
//
// An artifact's {T} cost is not held back by summoning sickness, which
// only creatures have (CR 302.6), so it can be activated the turn it
// arrives.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e2725411-a600-4ff3-94fb-5d52f5eda414",
		Name:         "Scepter of Insight",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}{U}, {T}: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    Plus(ManaCost("{3}{U}"), TapCost()),
			Effect:  ebDrawACard,
		}},
	})
}
