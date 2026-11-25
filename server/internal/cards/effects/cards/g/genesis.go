package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Genesis — Creature — Incarnation {4}{G}, 4/4 (Modern Horizons):
//
//	"At the beginning of your upkeep, if this creature is in your
//	 graveyard, you may pay {2}{G}. If you do, return target creature card
//	 from your graveyard to your hand."
//
// A trigger that works only from the graveyard (CR 113.6, InGraveyard),
// where "your" is the card's owner (CR 108.4). The "if" is an intervening
// "if" (CR 603.4): the ability triggers only while this object is in the
// graveyard (the zone it functions from), and as it resolves the same
// object must still be there, or nothing happens. The target is chosen as
// the ability goes on the stack, and may be Genesis itself. The payment is
// offered as the ability resolves (CR 603.5); if the target has left the
// graveyard by then, the ability does nothing for it (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "15afdd88-05fc-4335-8a55-c6ee702f3bef",
		Name:         "Genesis",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InGraveyard(effects.Targeting(effects.AtYourUpkeep("Genesis — you may pay {2}{G} to return target creature card from your graveyard to your hand",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.CardObjectStillInGraveyard(g, item) {
						return nil
					}
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{2}{G}",
						Question: "Genesis — pay {2}{G} to return target creature card from your graveyard to your hand?",
						Then:     genesisPaid,
					}.Apply(effects.NewContext(g, item))
				}), effects.TargetCreatureInYourGraveyard())),
		},
	})
}

var genesisPaid = effects.PayUnlessBranches("pay-unless/genesis-return-creature-card",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.ReturnFirstLegalGraveyardTargetToHand(ctx.Game, ctx.Item)
	}, nil)
