package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "6f5e6799-cb0f-4403-9c2a-69abeac08b79",
		Name:         "Watchful Automaton",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}: Scry 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Scry: 1},
			Cost:    effects.ManaCost("{2}{U}"),
			Effect:  effects.Do(effects.Scry{N: 1}),
		}},
	})
}
