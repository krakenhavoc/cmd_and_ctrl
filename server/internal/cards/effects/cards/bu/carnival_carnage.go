package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Carnival // Carnage — split card {B/R} // {2}{B}{R} (Ravnica Allegiance):
//
//	Carnival (Instant): "Carnival deals 1 damage to target creature or
//	 planeswalker and 1 damage to that permanent's controller."
//	Carnage (Sorcery): "Carnage deals 3 damage to target opponent. That player
//	 discards two cards."
//
// Carnival's two points of damage are dealt at once, as one damage instance
// (CR 120.2). The permanent's controller is read as the spell resolves; if the
// permanent is no longer a legal target the whole spell does nothing (CR 608.2b).
// Carnage's discard is the target player's own choice (CR 701.9a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b996ae86-3bf6-4e84-8302-6ec6edaa6ff0",
		Name:         "Carnival",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 1)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.FirstLegalTarget(ctx)
			if !ok {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(t.ID)
			if !ok {
				return nil
			}
			return ctx.Game.DamageInstanceForEffect(func() error {
				if err := (effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 1}).Apply(ctx); err != nil {
					return err
				}
				return effects.DealDamage{Source: ctx.Source(), Target: c.Controller, Amount: 1}.Apply(ctx)
			})
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "b996ae86-3bf6-4e84-8302-6ec6edaa6ff0#1",
		Name:         "Carnage",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.DealDamageToTheTarget(3)(ctx.Game, item); err != nil {
				return err
			}
			return effects.TargetPlayerDiscards(2)(ctx.Game, item)
		},
	})
}
