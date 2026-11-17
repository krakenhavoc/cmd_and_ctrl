package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inspired Sphinx — Creature — Sphinx {5}{U}{U}, 5/5 (Game Night,
// #5523):
//
//	"Flying
//	 When this creature enters, draw cards equal to the number of
//	 opponents you have.
//	 {3}{U}: Create a 1/1 colorless Thopter artifact creature token with
//	 flying."
//
// The draw counts your opponents as the trigger resolves (CR 608.2h), the
// Ethereal Investigator way: an opponent who has left the game is not
// counted. The Thopter is the shared token row Detective's Satchel makes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "61cceb14-9568-4a5e-9b56-07afa6407df2",
		Name:            "Inspired Sphinx",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Inspired Sphinx — draw cards equal to the number of opponents you have",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return DrawCards{N: len(ctx.Opponents())}.Apply(ctx)
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{U}: Create a 1/1 colorless Thopter artifact creature token with flying.",
			Cost:    ManaCost("{3}{U}"),
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1},
			Effect:  Do(CreateToken{Template: TokenCard("1/1 colorless Thopter artifact with flying"), N: 1}),
		}},
	})
}
