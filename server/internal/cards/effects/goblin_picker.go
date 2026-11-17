package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Picker — Creature — Goblin {1}{R}, 2/2 (Dominaria United):
//
//	"{R}, {T}, Discard a card: Draw a card."
//
// A looter whose discard is a cost (CR 118.8), paid before the draw, so a
// player with no card to discard cannot activate it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "efc0aa25-9ac2-4a25-956c-fff39ef536be",
		Name:         "Goblin Picker",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{R}, {T}, Discard a card: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{R}"), TapCost(), DiscardACard()),
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}
