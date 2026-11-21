package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cartographer's Survey — {3}{G} Sorcery (Innistrad: Crimson Vow, S82):
//
//	"Look at the top seven cards of your library. Put up to two land
//	 cards from among them onto the battlefield tapped. Put the rest on
//	 the bottom of your library in a random order."
//
// Elvish Rejuvenator's look-and-put (PutFromLibraryOntoBattlefield) with
// a ceiling of two instead of one, so the player may take none, one or
// two; the rest go to the bottom in a random order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e05029b4-552d-4691-a5c4-885284fb1636",
		Name:         "Cartographer's Survey",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Lands: 2},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.PutFromLibraryOntoBattlefield{
				Player:   item.Controller,
				Cards:    ctx.Game.LookAtTopOfLibraryForEffect(item.Controller, 7),
				Match:    effects.Land(),
				Max:      2,
				Optional: true,
				Tapped:   true,
				Label:    "Cartographer's Survey — put up to two land cards from among them onto the battlefield tapped",
				ThenTail: effects.PutRestOnBottomInRandomOrderTail,
			}.Apply(ctx)
		},
	})
}
