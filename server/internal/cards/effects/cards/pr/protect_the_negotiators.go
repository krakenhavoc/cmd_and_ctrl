package cardspr

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Protect the Negotiators — Instant {1}{U} (Dominaria United):
//
//	"Kicker {W} (You may pay an additional {W} as you cast this spell.)
//	 If this spell was kicked, create a 1/1 white Soldier creature token.
//	 Counter target spell unless its controller pays {1} for each
//	 creature you control."
//
// Printed order is resolution order (CR 608.2c): the kicked token is
// made first, so it counts toward the price. The creatures are counted
// once as the counter-unless runs. If the target spell is gone the
// spell does nothing, the token included (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:      "7ccd666a-0d56-4fe3-b16b-d035bf2ad420",
		Name:          "Protect the Negotiators",
		Purpose:       game.Purpose{Tokens: 1, CreatureTokens: 1},
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{W}")},
		Targets:       effects.TargetSpell("target spell"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.WasKicked() || len(ctx.LegalTargets()) == 0 {
					return nil
				}
				return effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier"), N: 1}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok {
					return nil
				}
				cost := "{" + strconv.Itoa(len(effects.CreaturesControlledByPlayer(ctx.Game, ctx.Controller()))) + "}"
				return effects.CounterUnlessPaid{
					StackID:  t.ID,
					Cost:     cost,
					Question: "Protect the Negotiators — pay " + cost + " or your spell is countered",
				}.Apply(ctx)
			}),
		),
	})
}
