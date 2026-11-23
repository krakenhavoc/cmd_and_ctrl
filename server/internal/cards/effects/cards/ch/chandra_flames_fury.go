package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chandra, Flame's Fury — Legendary Planeswalker {4}{R}{R}, loyalty 4
// (Jumpstart 2022, #5948):
//
//	"+1: Chandra deals 2 damage to any target.
//	 −2: Chandra deals 4 damage to target creature and 2 damage to that
//	 creature's controller.
//	 −8: Chandra deals 10 damage to target player and each creature that
//	 player controls."
//
// The −2 reads the creature's controller before it is dealt damage (CR
// 608.2h), so a creature that dies still has its controller hit. The −8
// is Chandra Nalaar's fan-out with a player-only target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "408966d2-08c0-4531-a3b4-e666840ee95c",
		Name:            "Chandra, Flame's Fury",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Chandra deals 2 damage to any target.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetAny(),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
				Effect:  effects.DamageToFirstLegalTarget(2),
			},
			{
				Label:   "−2: Chandra deals 4 damage to target creature and 2 damage to that creature's controller.",
				Cost:    effects.LoyaltyCost(-2),
				Targets: effects.TargetCreature("target creature"),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 4)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					t, ok := effects.FirstLegalTarget(ctx)
					if !ok {
						return nil
					}
					c, found := g.LookupCardForEffect(t.ID)
					if !found {
						return nil
					}
					controller := c.Controller
					if err := (effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 4}).Apply(ctx); err != nil {
						return err
					}
					if g.PlayerByIDForEffect(controller) == nil {
						return nil
					}
					return effects.DealDamage{Source: ctx.Source(), Target: controller, Amount: 2}.Apply(ctx)
				},
			},
			{
				Label:   "−8: Chandra deals 10 damage to target player and each creature that player controls.",
				Cost:    effects.LoyaltyCost(-8),
				Targets: effects.TargetPlayer("target player"),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 10)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					t, ok := effects.FirstLegalTarget(ctx)
					if !ok {
						return nil
					}
					ids := []uuid.UUID{t.ID}
					for _, c := range effects.MatchingBattlefield(ctx, effects.And(effects.Creature(), effects.ControlledBy(t.ID))) {
						ids = append(ids, c.InstanceID)
					}
					return effects.DealDamageToEachThen(ctx, ids, 10, effects.RecipientThen{})
				},
			},
		},
	})
}
