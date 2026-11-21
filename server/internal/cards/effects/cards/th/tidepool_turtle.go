package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tidepool Turtle — Creature — Turtle {3}{U}, 2/5 (Dominaria United, S82):
//
//	"{2}{U}: Scry 1."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ac2668a3-7a26-47b1-9ad4-56e23643de80",
		Name:         "Tidepool Turtle",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}: Scry 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Scry: 1},
			Cost:    effects.ManaCost("{2}{U}"),
			Effect:  effects.Do(effects.Scry{N: 1}),
		}},
	})
}
