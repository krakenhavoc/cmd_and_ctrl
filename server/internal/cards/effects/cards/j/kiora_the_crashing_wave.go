package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kiora, the Crashing Wave — Legendary Planeswalker — Kiora {2}{G}{U},
// loyalty 2:
//
//	"+1: Until your next turn, prevent all damage that would be dealt to
//	 and dealt by target permanent an opponent controls.
//	 −1: Draw a card. You may play an additional land this turn.
//	 −5: You get an emblem with 'At the beginning of your end step, create
//	 a 9/9 blue Kraken creature token.'"
//
// ADR 0108 §7, Delivery PR 7 (#1904): the +1 is Dovin's to-and-by row
// (one record, all damage, until your next turn). The −1 is Explore's
// draw and extra land play; the −5 is an emblem (#623) with one
// end-step trigger.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "58cc2097-f4b4-4695-9076-78666050cdb6",
		Name:            "Kiora, the Crashing Wave",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 2,
		Emblem: &effects.EmblemSpec{
			Label: "Kiora, the Crashing Wave emblem",
			Text:  "At the beginning of your end step, create a 9/9 blue Kraken creature token.",
			Triggered: []game.TriggeredAbility{
				effects.AtYourEndStep("Kiora, the Crashing Wave emblem — create a 9/9 blue Kraken",
					effects.Do(effects.CreateToken{Template: effects.TokenCard("9/9 blue Kraken"), N: 1})),
			},
		},
		Activated: []effects.ActivatedAbility{
			effects.UntilYourNextTurnToAndByRow("+1: Until your next turn, prevent all damage that would be dealt to and dealt by target permanent an opponent controls.",
				effects.LoyaltyCost(1)),
			{
				Label: "−1: Draw a card. You may play an additional land this turn.",
				Cost:  effects.LoyaltyCost(-1),
				// The extra land play waits for a prompt the draw asks (CR 608.2c, #4155).
				Effect: effects.Do(
					effects.DrawCards{N: 1},
					effects.Instruction(func(ctx *effects.Context) error {
						ctx.Game.GrantAdditionalLandPlayForEffect(ctx.Item.Controller, 1)
						return nil
					}),
				),
			},
			{
				Label:  "−5: You get an emblem with \"At the beginning of your end step, create a 9/9 blue Kraken creature token.\"",
				Cost:   effects.LoyaltyCost(-5),
				Effect: effects.Do(effects.CreateEmblem{}),
			},
		},
	})
}
