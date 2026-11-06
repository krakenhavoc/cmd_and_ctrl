package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Clachan Festival — Kindred Enchantment — Kithkin {2}{W} (Lorwyn
// Eclipsed, #4458):
//
//	"When this enchantment enters, create two 1/1 green and white Kithkin
//	 creature tokens.
//	 {4}{W}: Create a 1/1 green and white Kithkin creature token."
//
// The token is Kinbinding's 1/1 green and white Kithkin. The activated
// ability has no timing restriction and no tap, so it can be activated
// at instant speed as often as it can be paid for.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5396045d-744b-435e-8324-a1bfbb03449d",
		Name:         "Clachan Festival",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Clachan Festival — create two 1/1 green and white Kithkin creature tokens",
				Do(CreateToken{Template: TokenCard("1/1 green and white Kithkin"), N: 2})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{4}{W}: Create a 1/1 green and white Kithkin creature token.",
			Cost:    ManaCost("{4}{W}"),
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1},
			Effect:  Do(CreateToken{Template: TokenCard("1/1 green and white Kithkin"), N: 1}),
		}},
	})
}
