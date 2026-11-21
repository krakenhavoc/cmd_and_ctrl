package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moment of Truth — Instant {1}{U} (March of the Machine, S82 #5477):
//
//	"Look at the top three cards of your library. Put one of those cards
//	 into your hand, one into your graveyard, and one on the bottom of your
//	 library."
//
// A look, so only you see the cards. You choose the card for your hand, then
// the one for the graveyard; the last goes on the bottom. With fewer than
// three cards in the library the assignments are made in the printed order
// (hand, graveyard, bottom) with what is there.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ef61fb5c-2ea6-4f82-92c5-c4713dee45fa",
		Name:         "Moment of Truth",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			return effects.TakeFromLibraryToHand{
				Player: player,
				Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 3),
				Max:    1,
				Label:  "Moment of Truth — put one of those cards into your hand",
				Key:    effects.MomentOfTruthTake,
			}.Apply(ctx)
		},
	})
}
