package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Strategic Planning — Sorcery {1}{U} (Strixhaven Mystical Archive, #5503):
//
//	"Look at the top three cards of your library. Put one of them into
//	 your hand and the rest into your graveyard."
//
// Commune with Evil's look-and-take over three cards: a private look, a
// mandatory pick of one for the hand, the rest put into the graveyard
// (not milled or discarded). A shorter library offers what it has.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "02b5acf3-47cb-4d39-9307-e02656f1879b",
		Name:         "Strategic Planning",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			return effects.TakeFromLibraryToHand{
				Player: player,
				Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 3),
				Max:    1,
				Label:  "Strategic Planning — put one of them into your hand and the rest into your graveyard",
				Key:    effects.TakeRestGraveyard,
			}.Apply(ctx)
		},
	})
}
