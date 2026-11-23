package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Take Out the Trash — Instant {1}{R} (Bloomburrow, #4467):
//
//	"Take Out the Trash deals 3 damage to target creature or
//	 planeswalker. If you control a Raccoon, you may discard a card. If
//	 you do, draw a card."
//
// Plasma Cascade's shape: the damage, then Flaring Cinder's "you may
// discard a card. If you do, draw a card" (b39MayDiscardThenDraw), asked
// only if you control a Raccoon as the spell resolves, after the damage.
// "Raccoon" is read through the spell's text (CR 612.2). A spell whose
// only target has become illegal does not resolve (CR 608.2b), so it
// offers no rummage either.
//
// No simplification.
func init() {
	const question = "Take Out the Trash — you may discard a card; if you do, draw a card"
	effects.Register(effects.Spec{
		OracleID:     "e07700e7-218a-45a3-9ee4-de2cc25cdc02",
		Name:         "Take Out the Trash",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			if err := (effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 3}).Apply(ctx); err != nil {
				return err
			}
			if !effects.ControlsA(ctx.TextSubtype("Raccoon"))(ctx.Game, ctx.Controller()) {
				return nil
			}
			return effects.B39MayDiscardThenDraw(1, false, question, effects.DrawOneIfDiscarded)(ctx)
		},
	})
}
