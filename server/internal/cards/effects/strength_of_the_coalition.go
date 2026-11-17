package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:      "f0b24d59-b3f4-47e6-befd-40a25ad976e2",
		Name:          "Strength of the Coalition",
		Completeness:  CompletenessFull,
		Purpose:       ForTargets(BenefitUntilEOT(0, 2, 2)),
		Targets:       TargetCreature("target creature you control", YouControl()),
		OptionalCosts: []game.AdditionalCost{Kicker("{2}{W}")},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "Strength of the Coalition — +2/+2"}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				if !ctx.WasKicked() {
					return nil
				}
				return b11PutCountersOnEachCreatureYouControl(ctx.Game, ctx.Item, 1)
			}),
		),
	})
}
