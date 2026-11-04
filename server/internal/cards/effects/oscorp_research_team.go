package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Oscorp Research Team — Creature — Human Scientist {3}{U}, 1/5
// (Marvel's Spider-Man):
//
//	"{6}{U}: Draw two cards."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4d2e233c-0173-417f-82a0-1e692a400ae1",
		Name:         "Oscorp Research Team",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{6}{U}: Draw two cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{6}{U}"),
			Effect:  Do(DrawCards{N: 2}),
		}},
	})
}
