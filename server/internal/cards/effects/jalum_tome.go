package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "32d9fd98-142b-41d2-b8f0-40ae1d4cb991",
		Name:         "Jalum Tome",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Draw a card, then discard a card.",
			Purpose: game.Purpose{Draws: 1, Discards: 1, Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return b16DrawThenDiscard(g, item, 1)
			},
		}},
	})
}
