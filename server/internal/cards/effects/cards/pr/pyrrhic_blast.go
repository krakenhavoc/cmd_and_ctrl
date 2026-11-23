package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pyrrhic Blast — Instant {3}{R}:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Pyrrhic Blast deals damage equal to the sacrificed creature's power
//	 to any target. Draw a card."
//
// Fling with a card. The damage is the sacrificed creature's last-known
// power (the 2022-10-14 ruling, CR 608.2h), read off the payment record
// (ADR 0113 §1). A target that is illegal by resolution means the spell
// does not resolve at all, so no card is drawn (CR 608.2b).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "98ec58aa-8776-4a19-bff8-e288a9bd6bac",
		Name:           "Pyrrhic Blast",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("a creature", effects.Creature()),
		Targets:        effects.TargetAny(),
		// The draw waits for a prompt the damage asks (a CR 616 order; CR 608.2c, #4155).
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DamageEqualToSacrificedPower(ctx.Item, ctx)
			}),
			effects.DrawCards{N: 1},
		),
	})
}
