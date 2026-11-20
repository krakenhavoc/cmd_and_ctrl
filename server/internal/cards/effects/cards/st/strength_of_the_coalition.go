package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Strength of the Coalition — {G} Instant (Dominaria United):
//
//	"Kicker {2}{W} (You may pay an additional {2}{W} as you cast this
//	 spell.)
//	 Target creature you control gets +2/+2 until end of turn. If this
//	 spell was kicked, put a +1/+1 counter on each creature you control."
//
// The pump is followed by the counters (ResolveSteps). The counters go on
// the creatures you control as the spell resolves, one placement each.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "f0b24d59-b3f4-47e6-befd-40a25ad976e2",
		Name:          "Strength of the Coalition",
		Completeness:  effects.CompletenessFull,
		Purpose:       effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		Targets:       effects.TargetCreature("target creature you control", effects.YouControl()),
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{2}{W}")},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return effects.BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "Strength of the Coalition — +2/+2"}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.WasKicked() {
					return nil
				}
				return effects.B11PutCountersOnEachCreatureYouControl(ctx.Game, ctx.Item, 1)
			}),
		),
	})
}
