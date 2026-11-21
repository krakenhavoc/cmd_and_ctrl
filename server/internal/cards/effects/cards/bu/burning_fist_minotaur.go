package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Burning-Fist Minotaur — Creature — Minotaur Wizard {1}{R}, 2/1
// (Hour of Devastation, #5534):
//
//	"First strike
//	 {1}{R}, Discard a card: This creature gets +2/+0 until end of turn."
//
// First strike is a printed keyword. The pump's discard is a cost, paid as
// it is activated (CR 602.2b), on the shared thisGetsUntilEndOfTurn body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5460c76f-39dc-4db0-acf7-146d533d822b",
		Name:            "Burning-Fist Minotaur",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}, Discard a card: This creature gets +2/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 2}},
			Cost:    effects.Plus(effects.ManaCost("{1}{R}"), effects.DiscardACard()),
			Effect:  effects.ThisGetsUntilEndOfTurn(2, 0, "Burning-Fist Minotaur — +2/+0 until end of turn"),
		}},
	})
}
