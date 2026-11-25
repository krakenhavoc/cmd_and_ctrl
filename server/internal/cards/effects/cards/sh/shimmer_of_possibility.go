package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shimmer of Possibility — Sorcery {1}{U} (Ravnica Allegiance):
//
//	"Look at the top four cards of your library. Put one of them into your hand
//	 and the rest on the bottom of your library in a random order."
//
// A mandatory take of exactly one of the four, the rest bottomed at random
// (TakeRestBottomRandom).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "85e2b073-e953-46f9-a1d4-0dfe030b3751",
		Name:         "Shimmer of Possibility",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.TakeFromLibraryToHand{
				Player: ctx.Controller(),
				Cards:  ctx.Game.LookAtTopOfLibraryForEffect(ctx.Controller(), 4),
				Max:    1,
				Label:  "Shimmer of Possibility — put one of them into your hand",
				Key:    effects.TakeRestBottomRandom,
			}.Apply(ctx)
		},
	})
}
