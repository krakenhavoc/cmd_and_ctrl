package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Foundry Helix — Instant {1}{R}{W}:
//
//	"As an additional cost to cast this spell, sacrifice a permanent.
//	 Foundry Helix deals 4 damage to any target. If the sacrificed
//	 permanent was an artifact, you gain 4 life."
//
// "Was an artifact" is read off the sacrificed permanent as it last
// existed on the battlefield (CR 608.2h), from the payment record (ADR
// 0113 §1): a creature made an artifact by an effect counts. Any
// permanent pays, a land included.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:       "f4f66558-3c99-4488-ad2d-90626a922042",
		Name:           "Foundry Helix",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("a permanent", Permanent()),
		Targets:        TargetAny(),
		Purpose:        ForTargets(DamageToTarget(0, 4)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := damageToFirstTarget(4)(item, ctx); err != nil {
				return err
			}
			if !sacrificedHadCardType(ctx, "Artifact") {
				return nil
			}
			return GainLife{Amount: 4}.Apply(ctx)
		},
	})
}
