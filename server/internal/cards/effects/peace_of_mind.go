package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Peace of Mind — Enchantment {1}{W} (Exodus, #5626):
//
//	"{W}, Discard a card: You gain 3 life."
//
// Trading Post's discard-for-life row with a {W} cost and no tap: the
// discard is paid as the ability is activated (CR 602.2b, 601.2h), so it
// needs a card in hand, and it can be activated as often as its
// controller can pay.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4f8c5fd7-f280-4b0c-bb84-6ff9b258c50f",
		Name:         "Peace of Mind",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{W}, Discard a card: You gain 3 life.",
			Purpose: game.Purpose{LifeGain: 3, Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{W}"), DiscardACard()),
			Effect:  Do(GainLife{Amount: 3}),
		}},
	})
}
