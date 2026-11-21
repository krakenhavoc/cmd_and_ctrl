package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Killer Service — Enchantment {2}{G} (Murders at Karlov Manor Commander,
// #5368):
//
//	"When this enchantment enters, create a number of Food tokens equal
//	 to the number of opponents you have. ...
//	 At the beginning of your end step, you may pay {2} and sacrifice a
//	 token. If you do, create a 4/4 green Rhino Warrior creature token."
//
// The pay-and-sacrifice is offered only with a token to sacrifice.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0f30f32f-3c4c-4f73-8b2a-82045a40bda9",
		Name:         "Killer Service",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Killer Service — create a Food token for each opponent",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.CreateToken{Template: effects.FoodToken(), N: len(ctx.Opponents())}.Apply(ctx)
				}),
			effects.AtYourEndStep("Killer Service — you may pay {2} and sacrifice a token to create a 4/4 Rhino Warrior", effects.Cmdr2S18KillerServiceOffer),
		},
	})
}
