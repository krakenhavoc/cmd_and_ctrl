package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Garruk, Primal Hunter — Legendary Planeswalker — Garruk {2}{G}{G}{G},
// loyalty 3 (Magic 2012, #5573):
//
//	"+1: Create a 3/3 green Beast creature token.
//	 −3: Draw cards equal to the greatest power among creatures you
//	 control.
//	 −6: Create a 6/6 green Wurm creature token for each land you
//	 control."
//
// Three loyalty rows. The +1 is Garruk Wildspeaker's Beast; the -3 reads
// the greatest power among creatures you control as it resolves
// (b42GreatestPowerControlledBy, layered power); the -6 counts the lands
// you control as it resolves and makes that many Wurms.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "298a7e1a-b254-442e-bc1d-2a3e5296d849",
		Name:            "Garruk, Primal Hunter",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 3,
		Activated: []effects.ActivatedAbility{
			{
				Label:  "+1: Create a 3/3 green Beast creature token.",
				Cost:   effects.LoyaltyCost(1),
				Effect: effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 green Beast"), N: 1}),
			},
			{
				Label: "−3: Draw cards equal to the greatest power among creatures you control.",
				Cost:  effects.LoyaltyCost(-3),
				Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					return effects.DrawCards{N: effects.B42GreatestPowerControlledBy(ctx.Game, ctx.Controller())}.Apply(ctx)
				})),
			},
			{
				Label: "−6: Create a 6/6 green Wurm creature token for each land you control.",
				Cost:  effects.LoyaltyCost(-6),
				Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					n := effects.B02CountLandsControlledBy(ctx.Game, ctx.Controller())
					return effects.CreateToken{Template: effects.TokenCard("6/6 green Wurm"), N: n}.Apply(ctx)
				})),
			},
		},
	})
}
