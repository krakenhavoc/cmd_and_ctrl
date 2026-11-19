package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jungle Delver — Creature — Merfolk Warrior {G}, 1/1:
//
//	"{3}{G}: Put a +1/+1 counter on this creature."
//
// Ainok Bond-Kin's counter-on-this activation with a mana cost and no
// timing restriction.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2f299e69-d82b-48c9-b7f5-cf07cd9dced2",
		Name:         "Jungle Delver",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:  "{3}{G}: Put a +1/+1 counter on this creature.",
			Cost:   ManaCost("{3}{G}"),
			Effect: Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}),
			// A +1/+1 counter can outgrow damage in response.
			Purpose: game.Purpose{Answers: game.AnswerPump},
		}},
	})
}
