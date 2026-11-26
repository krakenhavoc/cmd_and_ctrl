package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Master the Way — Sorcery {3}{U}{R} (Khans of Tarkir):
//
//	"Draw a card. Master the Way deals damage to any target equal to the
//	 number of cards in your hand."
//
// The card is drawn first, so it counts; the hand is counted as the damage
// is dealt, with the spell itself already on the stack and not in hand.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "909a10f5-a607-474f-8b8b-0176fd840650",
		Name:         "Master the Way",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Draws: 1, Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			targets := ctx.LegalTargets()
			if len(targets) == 0 {
				return nil
			}
			if err := (effects.DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
				return err
			}
			return effects.DealDamage{Source: ctx.Source(), Target: targets[0].ID, Amount: effects.CardsInHand(ctx.Game, item.Controller)}.Apply(ctx)
		},
	})
}
