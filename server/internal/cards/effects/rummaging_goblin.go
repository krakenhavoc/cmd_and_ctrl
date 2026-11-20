package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rummaging Goblin — Creature — Goblin Rogue {2}{R}, 1/1 (Magic 2013, #5568):
//
//	"{T}, Discard a card: Draw a card."
//
// Goblin Picker's looter without the mana: the discard is a cost (CR
// 118.8), paid before the draw, so a player with no card to discard
// cannot activate it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2055eb91-ee36-4752-a6dc-581eaef335c8",
		Name:         "Rummaging Goblin",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Discard a card: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(TapCost(), DiscardACard()),
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}
