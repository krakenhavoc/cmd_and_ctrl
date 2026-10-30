package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Void Helix — Instant {4}{B}{B} (Marvel Super Heroes Commander,
// #4455): "Void Helix deals 5 damage to any target and you gain 5
// life." Lightning Helix's shape at five: the life gain waits for a
// prompt the damage asks (CR 608.2c), so it is its own step after the
// damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "82e52a09-ce6f-417d-8460-8806df63f0bf",
		Name:         "Void Helix",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      ForTargets(DamageToTarget(0, 5)),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				var target game.TargetRef
				if len(ctx.Item.Targets) > 0 {
					target = ctx.Item.Targets[0]
				}
				return DealDamage{Source: ctx.Source(), Target: target.ID, Amount: 5}.Apply(ctx)
			}),
			GainLife{Amount: 5},
		),
	})
}
