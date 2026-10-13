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
// DECLARED SIMPLIFICATION, weaker than printed and never stronger: the
// entry choice to exchange text boxes with another creature (CR 612,
// Exchange of Words's family) is not implemented, so Deadpool always
// keeps his own text box. The exchange is a "may", so declining it is
// a legal line in paper; what is missing is the other line. The seam is
// the text-changing effects row of docs/engine-seams.md (#2779): card
// abilities are catalog closures keyed by oracle ID rather than text,
// so there is nothing to swap yet.
//
// The upkeep loss is life loss on every upkeep, mandatory. "Each other
// player" is every player but the controller, teammates included.
func init() {
	Register(Spec{
		OracleID:     "4b96c22a-0d5b-44fd-b326-5f3ffcc3917b",
		Name:         "Deadpool, Trading Card",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"You can't exchange text boxes with another creature as Deadpool enters, so he always keeps his own.",
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
