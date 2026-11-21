package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "a2335149-d2db-48c1-9699-0df8ce12d4dd",
		Name:         "Death Grasp",
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      effects.ForTargets(effects.DamageXToTarget(0)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DamageXToTheTargetAndGainX(item, ctx, ctx.X())
		},
	})
}
