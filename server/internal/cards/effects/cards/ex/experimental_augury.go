package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Experimental Augury — Instant {1}{U} (Foundations Commander):
//
//	"Look at the top three cards of your library. Put one of them into
//	 your hand and the rest on the bottom of your library in any order.
//	 Proliferate."
//
// Impulse's look-and-take with three cards, whose "rest on the bottom in
// any order" ends in a proliferate: the take is registered with its own
// Then (cmdrB9AuguryTake), which orders the rest and hands the ordering
// the proliferate as its tail (cmdrB9ThenProliferate), so the proliferate
// waits for both the pick and the order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2ee08dcd-5558-45be-b619-8a550f3865ad",
		Name:         "Experimental Augury",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1, Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			return effects.TakeFromLibraryToHand{
				Player: player,
				Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 3),
				Max:    1,
				Label:  "Experimental Augury — put one of them into your hand",
				Key:    effects.CmdrB9AuguryTake,
			}.Apply(ctx)
		},
	})
}
