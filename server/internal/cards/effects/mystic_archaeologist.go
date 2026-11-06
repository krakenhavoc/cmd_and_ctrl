package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mystic Archaeologist — Creature — Human Wizard {1}{U}, 2/1
// (Foundations, #4465):
//
//	"{3}{U}{U}: Draw two cards."
//
// Oscorp Research Team's activated draw at a different cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b77dfe2a-ebc9-46b0-9134-2ecb2abdd8be",
		Name:         "Mystic Archaeologist",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}{U}{U}: Draw two cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 2},
			Cost:    ManaCost("{3}{U}{U}"),
			Effect:  Do(DrawCards{N: 2}),
		}},
	})
}
