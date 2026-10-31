package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Unbeatable Squirrel Girl — Legendary Creature — Squirrel Human Hero
// {1}{G}{G}{G}, 4/4 (Marvel Super Heroes, #4455):
//
//	"Do You Like Squirrels? — Whenever The Unbeatable Squirrel Girl enters
//	 or attacks, create a 1/1 green Squirrel creature token.
//	 I LOVE Squirrels! — {1}{G}{G}{G}: Create X 1/1 green Squirrel
//	 creature tokens, where X is the number of Squirrels you control."
//
// "Squirrel" is a word of the card's text, read through the text reader;
// X is counted on resolution and includes Squirrel Girl herself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "80312910-5d53-44f1-9982-e46dc7532abb",
		Name:         "The Unbeatable Squirrel Girl",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("The Unbeatable Squirrel Girl — create a 1/1 green Squirrel creature token",
				Do(CreateToken{Template: TokenCard("1/1 green Squirrel"), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{G}{G}{G}: Create X 1/1 green Squirrel creature tokens, where X is the number of Squirrels you control.",
			Cost:    ManaCost("{1}{G}{G}{G}"),
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1},
			Effect: func(g *game.Game, item *game.StackItem) error {
				squirrel := g.TextReaderForEffect(item).TextSubtype("Squirrel")
				x := 0
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == item.Controller && c.HasSubtype(squirrel) {
						x++
					}
				}
				if x == 0 {
					return nil
				}
				return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 green Squirrel"), N: x}.Apply(NewContext(g, item))
			},
		}},
	})
}
