package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Telekinetic Bonds — Enchantment {2}{U}{U}{U} (Judgment):
//
//	"Whenever a player discards a card, you may pay {1}{U}. If you do, you
//	 may tap or untap target permanent."
//
// Spirit Cairn's every-player discard trigger with a "may pay", and the
// target chosen as the trigger goes on the stack (CR 603.3d) whether or not
// the payment is made. The payment is asked as the trigger resolves; a yes
// is Rimewind Taskmage's tap-or-untap, whose choice includes leaving the
// permanent as it is.
//
// No simplification.
func init() {
	const name = "Telekinetic Bonds"
	effects.Register(effects.Spec{
		OracleID:     "708a76cc-d4ab-4dfd-98dd-815f5bc99364",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRestrict},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventDiscardCard, effects.AnyPlayer,
				name+" — you may pay {1}{U}; if you do, you may tap or untap target permanent",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}{U}",
						Question: name + " — pay {1}{U} to tap or untap target permanent?",
						Then:     telekineticBondsPaid,
					}.Apply(effects.NewContext(g, item))
				}), effects.TargetPermanent("target permanent")),
		},
	})
}

// telekineticBondsPaid is the paid branch: tap or untap the trigger's target.
// An on-disk key: never renamed, never reused.
var telekineticBondsPaid = effects.PayUnlessBranches("pay-unless/s82-old-b626-telekinetic-bonds-tap-or-untap",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.B38TapOrUntapTarget("Telekinetic Bonds")(ctx.Game, ctx.Item)
	}, nil)
