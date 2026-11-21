package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mercurial Chemister — Creature — Human Wizard {3}{U}{R}, 2/3:
//
//	"{U}, {T}: Draw two cards.
//	 {R}, {T}, Discard a card: This creature deals damage to target
//	 creature equal to the discarded card's mana value."
//
// ADR 0109 §8 (#1862): the damage is the discarded card's mana value,
// read off the payment record (Context.DiscardedManaValue, CR 400.7j)
// as the ability resolves. A land, or any card with mana value 0,
// deals no damage. The Chemister is the source, as it last existed if
// it has left the battlefield (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4243baf6-9ac2-4f87-988d-3ebf5e177f3b",
		Name:         "Mercurial Chemister",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{U}, {T}: Draw two cards.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{U}"), effects.TapCost()),
				Effect:  effects.Do(effects.DrawCards{N: 2}),
			},
			{
				Label:   "{R}, {T}, Discard a card: This creature deals damage to target creature equal to the discarded card's mana value.",
				Cost:    effects.Plus(effects.ManaCost("{R}"), effects.TapCost(), effects.DiscardACard()),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: ctx.DiscardedManaValue()}.Apply(ctx)
						}
					}
					return nil
				},
			},
		},
	})
}
