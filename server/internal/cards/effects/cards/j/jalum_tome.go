package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jalum Tome — Artifact — Book {3} (Antiquities, #5640):
//
//	"{2}, {T}: Draw a card, then discard a card."
//
// Isolinear Processor's loot row on its own: draw first, then the
// controller picks the discard from the hand that now holds the new card
// (b16DrawThenDiscard).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "32d9fd98-142b-41d2-b8f0-40ae1d4cb991",
		Name:         "Jalum Tome",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Draw a card, then discard a card.",
			Purpose: game.Purpose{Draws: 1, Discards: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.B16DrawThenDiscard(g, item, 1)
			},
		}},
	})
}
