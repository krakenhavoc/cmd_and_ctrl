package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jace's Archivist — Creature — Vedalken Wizard {1}{U}{U}, 2/2
// (EDHREC rank 1902):
//
//	"{U}, {T}: Each player discards their hand, then draws cards
//	 equal to the greatest number of cards a player discarded this
//	 way."
//
// A repeatable Windfall. Every player's hand is discarded before
// anyone draws (the whole-hand discard needs no choice, so it is the
// random-order discard, and discard payoffs queue while the ability
// is still resolving); the draw count is the largest hand that was
// pitched, taken from the discards themselves rather than from hand
// sizes read beforehand. Nobody draws when every hand was empty.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b6c8ac69-daa7-4e2e-a1d9-439731a81870",
		Name:         "Jace's Archivist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, {T}: Each player discards their hand, then draws cards equal to the greatest number discarded",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{U}"), effects.TapCost()),
			Effect:  effects.B17WheelToGreatestDiscard,
		}},
	})
}
