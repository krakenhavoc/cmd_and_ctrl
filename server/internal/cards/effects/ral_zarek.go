package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "75ed6946-e7b1-4bf8-a05c-62a2ba5a6c79",
		Name:            "Ral Zarek",
		Completeness:    CompletenessFull,
		StartingLoyalty: 4,
		Activated: []ActivatedAbility{
			{
				Label: "+1: Tap target permanent, then untap another target permanent.",
				Cost:  LoyaltyCost(1),
				Targets: Clauses(
					TargetPermanent("target permanent"),
					Distinct(TargetPermanent("another target permanent")),
				),
				Effect: Do(
					Instruction(func(ctx *Context) error {
						t, ok := ctx.ClauseTarget(0)
						if !ok || t.Kind != game.TargetCard {
							return nil
						}
						return TapTarget{Target: t.ID}.Apply(ctx)
					}),
					Instruction(func(ctx *Context) error {
						t, ok := ctx.ClauseTarget(1)
						if !ok || t.Kind != game.TargetCard {
							return nil
						}
						return UntapTarget{Target: t.ID}.Apply(ctx)
					}),
				),
			},
			{
				Label:   "−2: Ral Zarek deals 3 damage to any target.",
				Cost:    LoyaltyCost(-2),
				Purpose: ForTargets(DamageToTarget(0, 3)),
				Targets: TargetAny(),
				Effect:  DealDamageToTheTarget(3),
			},
			{
				Label: "−7: Flip five coins. Take an extra turn after this one for each coin that comes up heads.",
				Cost:  LoyaltyCost(-7),
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
					return TakeExtraTurn{N: heads}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
