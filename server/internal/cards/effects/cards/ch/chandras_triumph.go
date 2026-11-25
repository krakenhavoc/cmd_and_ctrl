package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chandra's Triumph — Instant {1}{R} (War of the Spark, #5520):
//
//	"Chandra's Triumph deals 3 damage to target creature or planeswalker an
//	 opponent controls. Chandra's Triumph deals 5 damage instead if you
//	 control a Chandra planeswalker."
//
// "A Chandra planeswalker" is the subtype on a planeswalker, read as the spell
// resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c0cc881c-b63b-44dd-8157-67680c65fc3d",
		Name:         "Chandra's Triumph",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		Targets: effects.TargetPermanent("target creature or planeswalker an opponent controls",
			effects.Or(effects.Creature(), effects.Planeswalker()), effects.OpponentControls()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			n := 3
			if effects.B42ControlsPlaneswalkerNamed(ctx.Game, ctx.Controller(), "Chandra") {
				n = 5
			}
			return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: n}.Apply(ctx)
		},
	})
}
