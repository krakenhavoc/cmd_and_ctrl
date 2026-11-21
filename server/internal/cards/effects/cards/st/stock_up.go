package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stock Up — Sorcery {2}{U}:
//
//	"Look at the top five cards of your library. Put two of them into
//	 your hand and the rest on the bottom of your library in any order."
//
// Dig Through Time's shape with five and two, and no delve: a private
// look, a mandatory take of two (fewer if the library holds fewer),
// and the ordered bottom (ADR 0088). Its purpose is Tutors: 2 for the
// same reason as Dig Through Time's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2251e34e-4ce8-4452-9dc6-d8cce8583996",
		Name:         "Stock Up",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 2},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			return effects.TakeFromLibraryToHand{
				Player: player,
				Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 5),
				Max:    2,
				Label:  "Stock Up — put two of them into your hand",
				Key:    effects.TakeRestBottomAny,
			}.Apply(ctx)
		},
	})
}
