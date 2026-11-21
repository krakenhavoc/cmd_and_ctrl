package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Recoil — Instant {1}{U}{B} (Dominaria Remastered, #5948):
//
//	"Return target permanent to its owner's hand. Then that player discards
//	 a card."
//
// "That player" is the permanent's owner (the player whose hand it returns
// to), read before it moves. The discard is the owner's own choice, and it
// happens whether or not the permanent reached the hand.
//
// No simplification.
var recoilDiscard = effects.CardTail("s82-gap-b141/recoil-discard", func(ctx *effects.Context, o game.TailOutcome) error {
	ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: o.Tail.Player, Source: o.Tail.Source, N: 1})
	return nil
})

func init() {
	effects.Register(effects.Spec{
		OracleID:     "1579bf86-4400-4467-88e8-1c5ba1209804",
		Name:         "Recoil",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target permanent"),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, found := ctx.Game.LookupCardForEffect(id)
			if !found {
				return nil
			}
			tail := ctx.Tail(recoilDiscard)
			tail.Player = c.Owner
			return effects.BounceToHand{Target: id, ThenTail: tail}.Apply(ctx)
		},
	})
}
