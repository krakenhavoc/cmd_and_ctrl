package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Faerie Formation — Creature — Faerie {4}{U}, 5/4 (Wilds of Eldraine
// Commander, #5368):
//
//	"Flying
//	 {3}{U}: Create a 1/1 blue Faerie creature token with flying. Draw a
//	 card."
//
// Both sentences are one activated ability's effect, in printed order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0366c2f6-6e78-4526-808c-fa7bace6006e",
		Name:         "Faerie Formation",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{U}: Create a 1/1 blue Faerie creature token with flying. Draw a card.",
			Purpose: game.Purpose{Draws: 1, Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker},
			Cost:    effects.ManaCost("{3}{U}"),
			Effect: effects.Do(
				effects.CreateToken{Template: effects.TokenCard("1/1 blue Faerie with flying"), N: 1},
				effects.DrawCards{N: 1},
			),
		}},
	})
}
