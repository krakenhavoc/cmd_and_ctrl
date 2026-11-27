package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lightning Dart — Instant {1}{R} (Invasion, S82 #5616):
//
//	"Lightning Dart deals 1 damage to target creature. If that creature is white
//	 or blue, Lightning Dart deals 4 damage to it instead."
//
// The colour is read as the spell resolves, through the layers, so a creature
// that became white or blue in response takes 4, and one that stopped being
// either takes 1. A target that has become illegal makes the spell do nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7bc53740-86e7-4e9d-aab6-b5ae62a42133",
		Name:         "Lightning Dart",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 4)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			n := 1
			if c, found := ctx.Game.LookupCardForEffect(t.ID); found && (c.HasColorWord("W") || c.HasColorWord("U")) {
				n = 4
			}
			return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: n}.Apply(ctx)
		},
	})
}
