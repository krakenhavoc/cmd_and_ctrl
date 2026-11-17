package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "5e414dc2-d0ff-43e5-9aa2-cd9d96de6d89",
		Name:            "Lithomantic Barrage",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		Targets:         TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		Purpose:         ForTargets(DamageToTarget(0, 5)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			n := 1
			if c, found := ctx.Game.LookupCardForEffect(t.ID); found {
				if OfColor("W")(ctx.Game, ctx.Controller(), c) || OfColor("U")(ctx.Game, ctx.Controller(), c) {
					n = 5
				}
			}
			return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: n}.Apply(ctx)
		},
	})
}
