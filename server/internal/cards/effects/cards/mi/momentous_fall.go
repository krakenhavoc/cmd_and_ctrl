package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Momentous Fall — Instant {2}{G}{G}:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 You draw cards equal to the sacrificed creature's power, then you
//	 gain life equal to its toughness."
//
// Both numbers come from ONE record: the sacrificed creature as it last
// existed on the battlefield (the 2010-06-15 ruling, CR 608.2h), read
// off the payment record (ADR 0113 §1). The draw happens first, then the
// life, as printed. A negative power draws nothing and a negative
// toughness gains nothing (CR 107.1b).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "6a5ce1b0-ac78-4a74-9eb7-4064e1687bf1",
		Name:           "Momentous Fall",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("a creature", effects.Creature()),
		OnResolve: effects.ResolveSteps(
			// The life gain waits for a prompt the draw asks (CR 608.2c, #4155). The
			// sacrificed creature is read from the payment record in each step.
			effects.Instruction(func(ctx *effects.Context) error {
				info, ok := ctx.SacrificedPermanent()
				if !ok {
					return nil
				}
				return effects.DrawCards{N: info.Power}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				info, ok := ctx.SacrificedPermanent()
				if !ok || info.Toughness <= 0 {
					return nil
				}
				return effects.GainLife{Amount: info.Toughness}.Apply(ctx)
			}),
		),
	})
}
