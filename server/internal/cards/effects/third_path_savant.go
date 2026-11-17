package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Third Path Savant — Creature — Human Wizard {2}{U}, 2/3 (The Brothers' War, #5482):
//
//	"{7}: Draw two cards."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6024cb2d-a6e7-4adc-8b34-7f4e0cac16ce",
		Name:         "Third Path Savant",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{7}: Draw two cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{7}"),
			Effect:  Do(DrawCards{N: 2}),
		}},
	})
}
