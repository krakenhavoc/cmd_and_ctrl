package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Plasma Cascade — Instant {2}{R} (Star Trek, #4973):
//
//	"Plasma Cascade deals 3 damage to any target. You may discard a
//	 card. If you do, draw a card."
//
// The damage comes first; then the caster is asked whether to discard,
// and draws only if a card was discarded (b39MayDiscardThenDraw, the
// Improvised Act shape). A target that has become illegal makes the
// whole spell do nothing (CR 608.2b), so there is no loot either.
//
// No simplification.
func init() {
	const question = "Plasma Cascade — you may discard a card; if you do, draw a card"
	effects.Register(effects.Spec{
		OracleID:     "46ce40e9-b7fb-4f40-a7ab-f8a2234d1d63",
		Name:         "Plasma Cascade",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			if err := (effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 3}).Apply(ctx); err != nil {
				return err
			}
			return effects.B39MayDiscardThenDraw(1, false, question, effects.DrawAsManyAsDiscarded)(ctx)
		},
	})
}
