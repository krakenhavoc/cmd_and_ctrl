package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Oread of Mountain's Blaze — Enchantment Creature — Nymph {1}{R}, 1/3 (Theros Beyond Death, #5514):
//
//	"{2}{R}, Discard a card: Draw a card."
//
// Famished Foragers' rummage: the discard is part of the cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ebbcbaa4-9dc2-4098-b6a2-393375d86f16",
		Name:         "Oread of Mountain's Blaze",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{R}, Discard a card: Draw a card.",
			Cost:    Plus(ManaCost("{2}{R}"), DiscardACard()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}
