package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Plasma Bolt — Sorcery {R} (Edge of Eternities):
//
//	"Plasma Bolt deals 2 damage to any target.
//	 Void — Plasma Bolt deals 3 damage instead if a nonland permanent
//	 left the battlefield this turn or a spell was warped this turn."
//
// Void (void.go) is read as the spell resolves. The damage is dealt once,
// at the size void gave it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ac5258c1-5b46-4892-a445-bfd56875f039",
		Name:         "Plasma Bolt",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 2)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			n := 2
			if effects.Void(ctx.Game) {
				n = 3
			}
			return effects.DealDamage{
				Source: ctx.Source(),
				Target: item.Targets[0].ID,
				Amount: n,
			}.Apply(ctx)
		},
	})
}
