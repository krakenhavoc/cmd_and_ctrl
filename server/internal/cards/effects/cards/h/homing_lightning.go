package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Homing Lightning — Instant {2}{R}{R} (Ravnica Remastered):
//
//	"Homing Lightning deals 4 damage to target creature and each other
//	 creature with the same name as that creature."
//
// The name is the target's layered name as the spell resolves. All the
// damage is one instruction (CR 615.8), dealt through
// DealDamageEachKeyedForEffect, target first. A target that has left makes
// the spell fizzle on its own (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6e7f8e25-49a9-49a0-9bf4-129439344550",
		Name:         "Homing Lightning",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 4)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			target, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			name := target.Effective().Name
			ids := []uuid.UUID{id}
			if name != "" {
				for _, c := range effects.MatchingBattlefield(ctx, effects.Creature()) {
					if c.InstanceID != id && c.Effective().Name == name {
						ids = append(ids, c.InstanceID)
					}
				}
			}
			return ctx.Game.DealDamageEachKeyedForEffect(item.SourceCardID, ids, 4, game.KeyedTail{})
		},
	})
}
