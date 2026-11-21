package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Liliana, Death Wielder — Legendary Planeswalker — Liliana {5}{B}{B},
// starting loyalty 5:
//
//	"+2: Put a -1/-1 counter on up to one target creature.
//	 −3: Destroy target creature with a -1/-1 counter on it.
//	 −10: Return all creature cards from your graveyard to the
//	 battlefield."
//
// The −10 returns every creature card in the controller's graveyard
// together, as the ability resolves, so each enters through the same
// replacement window and every enters trigger sees the whole batch.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e1b6cc44-41b6-43d0-9fe9-722cebb1c47b",
		Name:            "Liliana, Death Wielder",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+2: Put a -1/-1 counter on up to one target creature.",
				Cost:    effects.LoyaltyCost(2),
				Targets: effects.TargetCreature("up to one target creature").WithCount(0, 1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							return effects.AddCounter{Target: t.ID, Kind: game.CounterMinusOne, N: 1}.Apply(ctx)
						}
					}
					return nil
				},
			},
			{
				Label:   "−3: Destroy target creature with a -1/-1 counter on it.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetCreature("target creature with a -1/-1 counter on it", effects.CreatureWithCounter(game.CounterMinusOne)),
				Effect:  effects.DestroyFirstLegalCardTarget,
			},
			{
				Label: "−10: Return all creature cards from your graveyard to the battlefield.",
				Cost:  effects.LoyaltyCost(-10),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.ReturnFromGraveyardTogether{
						Targets: effects.GraveyardCardIDs(ctx, item.Controller, func(c game.Card) bool { return c.IsCreature() && !c.IsToken() }),
					}.Apply(ctx)
				},
			},
		},
	})
}
