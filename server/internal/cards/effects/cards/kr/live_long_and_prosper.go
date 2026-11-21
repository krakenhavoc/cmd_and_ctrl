package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Live Long and Prosper — Sorcery {5}{W}{U} (Star Trek, #4973):
//
//	"This spell can't be countered.
//	 Double your life total. Draw cards equal to the number of cards in
//	 your hand. You have no maximum hand size until your next turn."
//
// Three instructions in printed order, each waiting for any prompt the one
// before it asks (CR 608.2c). Doubling a life total is gaining (or, below
// zero, losing) life equal to it (CR 701.10e, doubleLifeTotal), so a
// "can't gain life" stops it. The hand is counted as the draw begins (CR
// 608.2h), with this spell on the stack and not in it. "Until your next
// turn" is Enter the Infinite's granted no-maximum entry with ADR 0063's
// until-your-next-turn duration.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ff6bc9f5-5a83-42d9-b1a4-f0979aed28ad",
		Name:            "Live Long and Prosper",
		Completeness:    effects.CompletenessFull,
		CantBeCountered: true,
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DoubleLifeTotal(ctx.Game, ctx.Item.SourceCardID, ctx.Controller())
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				n := 0
				if p := ctx.PlayerByID(ctx.Controller()); p != nil && p.Hand != nil {
					n = p.Hand.Size()
				}
				return effects.DrawCards{N: n}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return ctx.Game.GrantHandSizeForEffect(ctx.Controller(), game.HandSizeNoMaximum, 0,
					"Live Long and Prosper — no maximum hand size", ctx.Item.SourceCardID,
					effects.DurationUntilYourNextTurn(ctx, ctx.Controller()))
			}),
		),
	})
}
