package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lightning Helix — "Lightning Helix deals 3 damage to any target
// and you gain 3 life." Composition of two primitives (DealDamage +
// GainLife) in one card file — the canonical shape for S14's
// declarative DSL.
//
// Partial-target handling: if the original target slot is now
// illegal on resolve, the damage half no-ops (DealDamage silently
// returns nil when the target doesn't resolve to a player or
// battlefield card) and the life-gain half fires unconditionally —
// CR 608.2b's "partial illegal still resolves" case lets the
// caster still gain life even when the damage misses.
func init() {
	Register(Spec{
		OracleID: "800c258a-cfc4-4a54-a667-065ea8dea69e",
		Name:     "Lightning Helix",
		Targets:  TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			var target game.TargetRef
			if len(item.Targets) > 0 {
				target = item.Targets[0]
			}
			if err := (DealDamage{
				Source: ctx.Source(),
				Target: target.ID,
				Amount: 3,
			}).Apply(ctx); err != nil {
				return err
			}
			return GainLife{
				Player: ctx.Controller(),
				Amount: 3,
			}.Apply(ctx)
		},
	})
}
