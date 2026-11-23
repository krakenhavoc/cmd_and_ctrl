package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reckoner's Bargain — Instant {1}{B}:
//
//	"As an additional cost to cast this spell, sacrifice an artifact or
//	 creature. You gain life equal to the sacrificed permanent's mana
//	 value. Draw two cards."
//
// The life is the sacrificed permanent's mana value as it last existed
// on the battlefield (CR 608.2h), read off the payment record (ADR 0113
// §1); a token that is no copy is 0. Then the two cards, as printed.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "044ea111-ce87-49ab-98f7-ae8447f48ac5",
		Name:           "Reckoner's Bargain",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("an artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
		// The cards wait for a prompt the life gain asks (CR 608.2c, #4155).
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.GainLife{Amount: effects.SacrificedManaValue(ctx)}.Apply(ctx)
			}),
			effects.DrawCards{N: 2},
		),
	})
}
