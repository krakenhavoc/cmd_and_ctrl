package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Diabolic Vision — Sorcery {U}{B}:
//
//	"Look at the top five cards of your library. Put one of them into
//	 your hand and the rest on top of your library in any order."
//
// Impulse's look with the rest going back on top: a mandatory pick of one
// for the hand, then an order prompt over what is left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6b7b1a71-812d-4ff5-b38b-94d1cf634143",
		Name:         "Diabolic Vision",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1, Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			return effects.TakeFromLibraryToHand{
				Player: player,
				Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 5),
				Max:    1,
				Label:  "Diabolic Vision — put one of them into your hand and the rest on top of your library in any order",
				Key:    effects.Cmdr2B27TakeRestOnTop,
			}.Apply(ctx)
		},
	})
}
