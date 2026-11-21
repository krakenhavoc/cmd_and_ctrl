package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lithomantic Barrage — Sorcery {R} (March of the Machine):
//
//	"This spell can't be countered.
//	 Lithomantic Barrage deals 1 damage to target creature or
//	 planeswalker. It deals 5 damage instead if that target is white
//	 and/or blue."
//
// The target's colour is read once, as the spell resolves (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5e414dc2-d0ff-43e5-9aa2-cd9d96de6d89",
		Name:            "Lithomantic Barrage",
		Completeness:    effects.CompletenessFull,
		CantBeCountered: true,
		Targets:         effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		Purpose:         effects.ForTargets(effects.DamageToTarget(0, 5)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			n := 1
			if c, found := ctx.Game.LookupCardForEffect(t.ID); found {
				if effects.OfColor("W")(ctx.Game, ctx.Controller(), c) || effects.OfColor("U")(ctx.Game, ctx.Controller(), c) {
					n = 5
				}
			}
			return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: n}.Apply(ctx)
		},
	})
}
