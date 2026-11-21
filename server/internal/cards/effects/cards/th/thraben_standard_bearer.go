package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "679d5a1d-8da8-4142-924b-97b359f55309",
		Name:         "Thraben Standard Bearer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}, {T}, Discard a card: Create a 1/1 white Human Soldier creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.Plus(effects.ManaCost("{1}{W}"), effects.TapCost(), effects.DiscardACard()),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Human Soldier"), N: 1}),
		}},
	})
}
