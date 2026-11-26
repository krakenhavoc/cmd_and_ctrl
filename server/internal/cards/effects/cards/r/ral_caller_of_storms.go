package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ral, Caller of Storms — Legendary Planeswalker — Ral {4}{U}{R}, loyalty 4
// (Guilds of Ravnica, #5525):
//
//	"+1: Draw a card.
//	 −2: Ral deals 3 damage divided as you choose among one, two, or three
//	     targets.
//	 −7: Draw seven cards. Ral deals 7 damage to each creature your opponents
//	     control."
//
// The +1 is a plain draw. The −2 is Inferno Titan's divided damage over one to
// three targets (any target), the division chosen as the ability goes on the
// stack (CR 601.2d); a target that left or became illegal is skipped (CR 608.2b)
// and its share is not redistributed. The −7 draws seven, then Ral deals 7 damage
// to each creature an opponent controls as one damage instruction, Blazing
// Volley's sweep.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d7b10518-b6dc-4a39-881a-edd346034f28",
		Name:            "Ral, Caller of Storms",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Draw a card.",
				Cost:    effects.LoyaltyCost(1),
				Purpose: game.Purpose{Draws: 1},
				Effect:  effects.Do(effects.DrawCards{N: 1}),
			},
			{
				Label:   "−2: Ral deals 3 damage divided as you choose among one, two, or three targets.",
				Cost:    effects.LoyaltyCost(-2),
				Targets: effects.TargetAny().WithCount(1, 3).Dividing(effects.Divide(3)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.DealDividedDamage(effects.NewContext(g, item))
				},
			},
			{
				Label:   "−7: Draw seven cards. Ral deals 7 damage to each creature your opponents control.",
				Cost:    effects.LoyaltyCost(-7),
				Purpose: game.Purpose{Draws: 7, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 7, OpponentsOnly: true}},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.DrawCards{N: 7}).Apply(ctx); err != nil {
						return err
					}
					return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.OpponentControls()), 7)
				},
			},
		},
	})
}
