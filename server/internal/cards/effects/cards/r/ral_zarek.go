package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ral Zarek — Legendary Planeswalker — Ral {2}{U}{R}, loyalty 4 (Ravnica
// Remastered):
//
//	"+1: Tap target permanent, then untap another target permanent.
//	 −2: Ral Zarek deals 3 damage to any target.
//	 −7: Flip five coins. Take an extra turn after this one for each coin
//	 that comes up heads."
//
// The +1's two targets are distinct clauses (CR 601.2c: "another" is a
// different object); the tap and the untap each do nothing if their target
// has left (CR 608.2b). The −7 flips through the engine's random draw (as
// Ral Zarek, Guest Lecturer does) and queues one extra turn per heads.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "75ed6946-e7b1-4bf8-a05c-62a2ba5a6c79",
		Name:            "Ral Zarek",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Activated: []effects.ActivatedAbility{
			{
				Label: "+1: Tap target permanent, then untap another target permanent.",
				Cost:  effects.LoyaltyCost(1),
				Targets: effects.Clauses(
					effects.TargetPermanent("target permanent"),
					effects.Distinct(effects.TargetPermanent("another target permanent")),
				),
				Effect: effects.Do(
					effects.Instruction(func(ctx *effects.Context) error {
						t, ok := ctx.ClauseTarget(0)
						if !ok || t.Kind != game.TargetCard {
							return nil
						}
						return effects.TapTarget{Target: t.ID}.Apply(ctx)
					}),
					effects.Instruction(func(ctx *effects.Context) error {
						t, ok := ctx.ClauseTarget(1)
						if !ok || t.Kind != game.TargetCard {
							return nil
						}
						return effects.UntapTarget{Target: t.ID}.Apply(ctx)
					}),
				),
			},
			{
				Label:   "−2: Ral Zarek deals 3 damage to any target.",
				Cost:    effects.LoyaltyCost(-2),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 3)),
				Targets: effects.TargetAny(),
				Effect:  effects.DealDamageToTheTarget(3),
			},
			{
				Label: "−7: Flip five coins. Take an extra turn after this one for each coin that comes up heads.",
				Cost:  effects.LoyaltyCost(-7),
				Effect: func(g *game.Game, item *game.StackItem) error {
					faces, err := g.FlipCoinsForEffect(game.RandomDraw{Player: item.Controller, Source: item.SourceCardID}, 5)
					if err != nil {
						return err
					}
					heads := 0
					for _, h := range faces {
						if h {
							heads++
						}
					}
					if heads == 0 {
						return nil
					}
					return effects.TakeExtraTurn{N: heads}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
