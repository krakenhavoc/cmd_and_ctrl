package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Worldly Counsel — Instant {1}{U} (Conflux):
//
//	"Domain — Look at the top X cards of your library, where X is the number
//	 of basic land types among lands you control. Put one of those cards into
//	 your hand and the rest on the bottom of your library in any order."
//
// Consult the Star Charts' look-and-take with the domain count (DomainCount,
// #5666), read once as the spell resolves; the pick is mandatory and the rest
// go to the bottom in an order the caster chooses (TakeRestBottomAny). With no
// basic land type among your lands X is 0 and nothing happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "079b1f07-6910-491b-83ba-5ec41260f32d",
		Name:         "Worldly Counsel",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			x := effects.DomainCount(ctx.Game, player)
			if x == 0 {
				return nil
			}
			return effects.TakeFromLibraryToHand{
				Player: player,
				Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, x),
				Max:    1,
				Label:  "Worldly Counsel — put one of those cards into your hand",
				Key:    effects.TakeRestBottomAny,
			}.Apply(ctx)
		},
	})
}
