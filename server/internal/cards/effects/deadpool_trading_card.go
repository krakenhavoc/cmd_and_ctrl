package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deadpool, Trading Card — Legendary Creature — Mutant Mercenary Hero
// {2}{B}{R}, 5/3:
//
//	"As Deadpool enters, you may exchange his text box and another
//	 creature's.
//	 At the beginning of your upkeep, you lose 3 life.
//	 {3}, Sacrifice this creature: Each other player draws a card."
//
// The entry exchange is CR 612.5's text-box exchange (ADR 0160): as he
// enters, his controller may pick another creature, any controller's,
// and the two swap rules text indefinitely. It is chosen before he
// lands and made before his enters event, so an enters trigger in the
// text he takes triggers for him (CR 603.6a). The creature he exchanged
// with keeps its own name, types and P/T and gets his upkeep loss and
// his sacrifice ability; "Deadpool" in that text means that creature
// (CR 201.5b). Declining is a legal answer.
//
// The upkeep loss is life loss on every upkeep, mandatory. "Each other
// player" is every player but the controller, teammates included.
func init() {
	Register(Spec{
		OracleID:     "4b96c22a-0d5b-44fd-b326-5f3ffcc3917b",
		Name:         "Deadpool, Trading Card",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersExchangingTextBoxes("Deadpool, Trading Card", anotherCreatureOnBattlefield),
		},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Deadpool — you lose 3 life", func(g *game.Game, item *game.StackItem) error {
				return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -3)
			}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}, Sacrifice this creature: Each other player draws a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{3}"), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, opp := range ctx.Opponents() {
					if err := (DrawCards{Player: opp, N: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
