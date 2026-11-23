package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pigment Storm — Sorcery {3}{R}{R} (Foundations Jumpstart):
//
//	"Pigment Storm deals 5 damage to target creature. Excess damage is
//	 dealt to that creature's controller instead."
//
// Ram Through's split, always on: the creature is dealt lethal damage
// (toughness less damage already marked, floored at zero) and the rest of
// the five goes to its controller, both in one damage instance from the
// spell. A target gone by resolution takes nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "89abdf4c-ba9d-4602-ae50-cf79c68ea08d",
		Name:         "Pigment Storm",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 5)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			victim, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			lethal := max(victim.CurrentToughness()-victim.DamageMarked, 0)
			toCreature := min(5, lethal)
			toController := 5 - toCreature
			return ctx.Game.DamageInstanceForEffect(func() error {
				if toCreature > 0 {
					if err := (effects.DealDamage{Source: item.SourceCardID, Target: id, Amount: toCreature}).Apply(ctx); err != nil {
						return err
					}
				}
				if toController > 0 {
					return effects.DealDamage{Source: item.SourceCardID, Target: victim.Controller, Amount: toController}.Apply(ctx)
				}
				return nil
			})
		},
	})
}
