package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "0366c2f6-6e78-4526-808c-fa7bace6006e",
		Name:         "Faerie Formation",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}{U}: Create a 1/1 blue Faerie creature token with flying. Draw a card.",
			Purpose: game.Purpose{Draws: 1, Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker},
			Cost:    ManaCost("{3}{U}"),
			Effect: Do(
				CreateToken{Template: TokenCard("1/1 blue Faerie with flying"), N: 1},
				DrawCards{N: 1},
			),
		}},
	})
}
