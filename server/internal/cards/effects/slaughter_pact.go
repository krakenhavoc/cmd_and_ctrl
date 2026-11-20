package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Slaughter Pact — Instant {0}, black (Future Sight, #5594):
//
//	"Destroy target nonblack creature.
//	 At the beginning of your next upkeep, pay {2}{B}. If you don't, you
//	 lose the game."
//
// Hand of Death's destroy, then Pact of Negation's upkeep debt. The debt
// is owed whether or not the target was still there to destroy. The card
// is black by its colour indicator, which rides game.Card.Colors from
// Scryfall.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "67c4b1d5-757c-46cf-8bb1-9889d9ed1a1f",
		Name:         "Slaughter Pact",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target nonblack creature", NonBlack()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := destroyTheTargetPermanent(item, ctx); err != nil {
				return err
			}
			return futB1OwePact(ctx, "Slaughter Pact", "{2}{B}")
		},
	})
}
