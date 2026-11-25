package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vraska, Regal Gorgon — Legendary Planeswalker — Vraska {5}{B}{G}, loyalty 5
// (Guilds of Ravnica, #5525):
//
//	"+2: Put a +1/+1 counter on up to one target creature. That creature gains
//	     menace until end of turn.
//	 −3: Destroy target creature.
//	 −10: For each creature card in your graveyard, put a +1/+1 counter on each
//	      creature you control."
//
// The +2 targets up to one creature, and the menace goes to that same creature
// if it is still a legal target. The −3 is an ordinary destroy. The −10 counts the
// creature cards in your graveyard once, as it resolves, and puts that many
// +1/+1 counters on each creature you control as one placement per creature
// (Nykthos Paragon's B28PutCountersOnEachCreatureYouControl, so a counter
// replacement applies once), the set of creatures fixed first.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7616e834-4852-40c8-9de6-d89f96119f12",
		Name:            "Vraska, Regal Gorgon",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+2: Put a +1/+1 counter on up to one target creature. That creature gains menace until end of turn.",
				Cost:    effects.LoyaltyCost(2),
				Targets: effects.TargetCreature("up to one target creature").WithCount(0, 1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (effects.AddCounter{Target: t.ID, Kind: "+1/+1", N: 1}).Apply(ctx); err != nil {
							return err
						}
						return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"menace"}, Label: "Vraska, Regal Gorgon — menace"}.Apply(ctx)
					}
					return nil
				},
			},
			{
				Label:   "−3: Destroy target creature.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.DestroyTheTargetPermanent(item, effects.NewContext(g, item))
				},
			},
			{
				Label: "−10: For each creature card in your graveyard, put a +1/+1 counter on each creature you control.",
				Cost:  effects.LoyaltyCost(-10),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					n := len(effects.GraveyardCardIDs(ctx, item.Controller, func(c game.Card) bool { return c.IsCreature() && !c.IsToken() }))
					return effects.B28PutCountersOnEachCreatureYouControl(g, item, n)
				},
			},
		},
	})
}
