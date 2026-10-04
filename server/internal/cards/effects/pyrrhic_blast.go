package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:       "98ec58aa-8776-4a19-bff8-e288a9bd6bac",
		Name:           "Pyrrhic Blast",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("a creature", Creature()),
		Targets:        TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := damageEqualToSacrificedPower(item, ctx); err != nil {
				return err
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
	})
}
