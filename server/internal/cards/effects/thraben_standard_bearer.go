package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thraben Standard Bearer — Creature — Human Soldier {W}, 1/1 (Eldritch
// Moon, #5541):
//
//	"{1}{W}, {T}, Discard a card: Create a 1/1 white Human Soldier creature
//	 token."
//
// Cryptbreaker's first ability with the token table's Human Soldier row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "679d5a1d-8da8-4142-924b-97b359f55309",
		Name:         "Thraben Standard Bearer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}, {T}, Discard a card: Create a 1/1 white Human Soldier creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    Plus(ManaCost("{1}{W}"), TapCost(), DiscardACard()),
			Effect:  Do(CreateToken{Template: TokenCard("1/1 white Human Soldier"), N: 1}),
		}},
	})
}
