package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Third Path Savant — Creature — Human Wizard {2}{U}, 2/3 (The Brothers' War, #5482):
//
//	"{7}: Draw two cards."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6024cb2d-a6e7-4adc-8b34-7f4e0cac16ce",
		Name:         "Third Path Savant",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{7}: Draw two cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{7}"),
			Effect:  effects.Do(effects.DrawCards{N: 2}),
		}},
	})
}
