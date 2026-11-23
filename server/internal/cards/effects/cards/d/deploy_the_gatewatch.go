package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deploy the Gatewatch — Sorcery {4}{W}{W} (Commander Masters, #5948):
//
//	"Look at the top seven cards of your library. Put up to two planeswalker
//	 cards from among them onto the battlefield. Put the rest on the bottom
//	 of your library in a random order."
//
// Cartographer's Survey's look-and-put (PutFromLibraryOntoBattlefield) over
// planeswalker cards, untapped, with the same ceiling of two: none, one or
// two may be taken, and the rest go to the bottom in a random order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e17ea7eb-362e-4c61-97cd-537f3921bced",
		Name:         "Deploy the Gatewatch",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.PutFromLibraryOntoBattlefield{
				Player:   item.Controller,
				Cards:    ctx.Game.LookAtTopOfLibraryForEffect(item.Controller, 7),
				Match:    effects.Planeswalker(),
				Max:      2,
				Optional: true,
				Label:    "Deploy the Gatewatch — put up to two planeswalker cards from among them onto the battlefield",
				ThenTail: effects.PutRestOnBottomInRandomOrderTail,
			}.Apply(ctx)
		},
	})
}
