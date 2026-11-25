package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drainpipe Vermin — Creature — Rat {B}, 1/1 (Jumpstart, #5510):
//
//	"When this creature dies, you may pay {B}. If you do, target player
//	 discards a card."
//
// The target player is chosen as the trigger goes on the stack (CR 603.3d);
// the payment is asked as it resolves (CR 118.12). Paid, the target player
// (if still legal, CR 608.2b) chooses and discards a card (CR 701.9a).
//
// No simplification.
func init() {
	const label = "Drainpipe Vermin — you may pay {B}; if you do, target player discards a card"
	effects.Register(effects.Spec{
		OracleID:     "b6d39905-e074-4517-b806-d081d1069f74",
		Name:         "Drainpipe Vermin",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisDies(label, func(g *game.Game, item *game.StackItem) error {
				return effects.MayPay{
					Chooser:  item.Controller,
					Cost:     "{B}",
					Question: "Drainpipe Vermin — pay {B} to make target player discard a card?",
					Then:     drainpipeVerminPaid,
				}.Apply(effects.NewContext(g, item))
			}), effects.TargetPlayer("target player")),
		},
	})
}

// drainpipeVerminPaid is the paid branch: the target player discards a
// card. An on-disk key: never renamed, never reused.
var drainpipeVerminPaid = effects.PayUnlessBranches("pay-unless/drainpipe-vermin-discard",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.TargetPlayerDiscards(1)(ctx.Game, ctx.Item)
	}, nil)
