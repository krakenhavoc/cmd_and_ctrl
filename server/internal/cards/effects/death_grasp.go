package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Death Grasp — Sorcery {X}{W}{B} (Apocalypse, #5614):
//
//	"Death Grasp deals X damage to any target. You gain X life."
//
// Blaze's X damage, then the caster gains X. The life is the announced
// X, not the damage dealt, so it is gained in full when the damage is
// prevented. A target that became illegal stops the whole spell (CR
// 608.2b: its only target is gone), so no life is gained then.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a2335149-d2db-48c1-9699-0df8ce12d4dd",
		Name:         "Death Grasp",
		XMatters:     true,
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      ForTargets(DamageXToTarget(0)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			if err := (DealDamage{Source: ctx.Source(), Target: item.Targets[0].ID, Amount: ctx.X()}).Apply(ctx); err != nil {
				return err
			}
			return GainLife{Player: ctx.Controller(), Amount: ctx.X()}.Apply(ctx)
		},
	})
}
