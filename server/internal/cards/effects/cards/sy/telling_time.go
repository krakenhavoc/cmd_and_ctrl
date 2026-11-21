package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Telling Time — Instant {1}{U}:
//
//	"Look at the top three cards of your library. Put one of those cards
//	 into your hand, one on top of your library, and one on the bottom of
//	 your library."
//
// Impulse's look, with a mandatory pick of exactly one for the hand, and
// then the rest split by Cream of the Crop's exact-count lane: one on top,
// the other on the bottom, the player's choice of which. With a library
// of fewer than three cards the cards that are there are placed as far as
// they go (CR 608.2: do as much as you can).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "192f3c8a-1120-4f7f-ae0f-fe9bde59cd08",
		Name:         "Telling Time",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1, Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			return effects.TakeFromLibraryToHand{
				Player: player,
				Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 3),
				Max:    1,
				Label:  "Telling Time — put one of those cards into your hand, one on top of your library, and one on the bottom",
				Key:    effects.Cmdr2B27TakeRestTopAndBottom,
			}.Apply(ctx)
		},
	})
}
