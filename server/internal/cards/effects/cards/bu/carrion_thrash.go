package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Carrion Thrash — {2}{B}{R}{G} Creature — Lizard Warrior 4/4 (Shards of Alara,
// S82 OLD b516):
//
//	"When this creature dies, you may pay {2}. If you do, return another
//	 target creature card from your graveyard to your hand."
//
// Dutiful Attendant's target and return behind Drainpipe Vermin's optional
// payment: the target is chosen as the trigger goes on the stack, the payment
// is asked as it resolves, and "another" keeps the Thrash itself out.
//
// No simplification.
func init() {
	const label = "Carrion Thrash — you may pay {2}; if you do, return another target creature card from your graveyard to your hand"
	effects.Register(effects.Spec{
		OracleID:     "89599a64-19fc-47ad-a116-3ce0a10fa546",
		Name:         "Carrion Thrash",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisDies(label, func(g *game.Game, item *game.StackItem) error {
				return effects.MayPay{
					Chooser:  item.Controller,
					Cost:     "{2}",
					Question: "Carrion Thrash — pay {2} to return another target creature card from your graveyard to your hand?",
					Then:     carrionThrashPaid,
				}.Apply(effects.NewContext(g, item))
			}), effects.Another(effects.TargetCardInGraveyard("another target creature card from your graveyard", effects.Creature(), effects.YouOwn()))),
		},
	})
}

var carrionThrashPaid = effects.PayUnlessBranches("pay-unless/old-b516-carrion-thrash-return",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.ReturnTargetFromGraveyardToHand(ctx.Item, ctx)
	}, nil)
