package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nicol Bolas, the Ravager // Nicol Bolas, the Arisen — {1}{U}{B}{R}
// Legendary Creature — Elder Dragon 4/4 // Legendary Planeswalker — Bolas,
// loyalty 7:
//
//	"Flying
//	 When Nicol Bolas enters, each opponent discards a card.
//	 {4}{U}{B}{R}: Exile Nicol Bolas, then return him to the battlefield
//	 transformed under his owner's control. Activate only as a sorcery."
//	"+2: Draw two cards.
//	 −3: Nicol Bolas deals 10 damage to target creature or planeswalker.
//	 −4: Put target creature or planeswalker card from a graveyard onto the
//	 battlefield under your control.
//	 −12: Exile all but the bottom card of target player's library."
//
// No simplification.
func init() {
	const id = "55e4b27e-5447-4fc2-8cae-a03e344600c6"
	effects.Register(effects.Spec{
		OracleID:        id,
		Name:            "Nicol Bolas, the Ravager",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.Self, "Nicol Bolas, the Ravager — each opponent discards a card", effects.EachOpponentDiscardsOne),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{4}{U}{B}{R}: Exile Nicol Bolas, then return him to the battlefield transformed under his owner's control. Activate only as a sorcery.",
			Cost:         effects.ManaCost("{4}{U}{B}{R}"),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ExileAndReturnTransformed{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})

	effects.Register(effects.Spec{
		OracleID:        id + "#1",
		Name:            "Nicol Bolas, the Arisen",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 7,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+2: Draw two cards.",
				Cost:    effects.LoyaltyCost(2),
				Purpose: game.Purpose{Draws: 2},
				Effect:  effects.Do(effects.DrawCards{N: 2}),
			},
			{
				Label:   "−3: Nicol Bolas deals 10 damage to target creature or planeswalker.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 10)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					t, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.DealDamage{Source: item.SourceCardID, Target: t, Amount: 10}.Apply(ctx)
				},
			},
			{
				Label:   "−4: Put target creature or planeswalker card from a graveyard onto the battlefield under your control.",
				Cost:    effects.LoyaltyCost(-4),
				Targets: effects.TargetCardInGraveyard("target creature or planeswalker card from a graveyard", effects.Or(effects.Creature(), effects.Planeswalker())),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.ReturnFromGraveyardTogether{Targets: effects.LegalTargetCardIDs(ctx), Controller: ctx.Controller()}.Apply(ctx)
				},
			},
			{
				Label:   "−12: Exile all but the bottom card of target player's library.",
				Cost:    effects.LoyaltyCost(-12),
				Targets: effects.TargetPlayer("target player"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					p, ok := effects.FirstLegalPlayerTarget(ctx)
					if !ok {
						return nil
					}
					pl := g.PlayerByIDForEffect(p)
					if pl == nil || pl.Library == nil || len(pl.Library.Cards) < 2 {
						return nil
					}
					return g.ExileTopKeyedForEffect(p, item.Controller, len(pl.Library.Cards)-1, game.KeyedTail{})
				},
			},
		},
	})
}
