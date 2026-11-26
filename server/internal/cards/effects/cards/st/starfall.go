package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Starfall — Instant {4}{R} (Journey into Nyx, S82 OLD, #5558):
//
//	"Starfall deals 3 damage to target creature. If that creature is an
//	 enchantment, Starfall deals 3 damage to that creature's controller."
//
// The creature is read as it resolves: an enchantment creature (judged on the
// layered view) also costs its controller 3. A target that is gone or has
// stopped being a legal target takes nothing and its controller is not hit
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b2c840a7-24b4-4ae3-a13f-ed524e18d488",
		Name:         "Starfall",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.FirstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			card, ok := ctx.Game.LookupCardForEffect(t.ID)
			if !ok {
				return nil
			}
			controller := card.Controller
			enchantment := card.IsEnchantment()
			if err := (effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 3}).Apply(ctx); err != nil {
				return err
			}
			if enchantment {
				return effects.DealDamage{Source: item.SourceCardID, Target: controller, Amount: 3}.Apply(ctx)
			}
			return nil
		},
	})
}
