package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Carrion Locust — Creature — Insect Horror {2}{B}, 2/1 (The Brothers' War,
// #5482):
//
//	"Flying
//	 When this creature enters, exile target card from an opponent's
//	 graveyard. If it was a creature card, that player loses 1 life."
//
// The card's kind is read before the exile and rides the exile's keyed
// tail, which runs only once the card has reached exile (CR 614); "that
// player" is its owner, who has the graveyard it was in. A target that left in response leaves the trigger with no legal
// target, so nothing is exiled and nobody loses life (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "04e5ecda-9646-4bc3-addf-cca471553fbe",
		Name:            "Carrion Locust",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Carrion Locust — exile target card from an opponent's graveyard; if it was a creature card, that player loses 1 life",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							return effects.ExileThenIfItWas{Target: t.ID, Was: effects.WasCreatureCard, Keyed: carrionLocustLoseLife}.Apply(ctx)
						}
						return nil
					}),
				effects.TargetCardInGraveyard("target card from an opponent's graveyard", effects.Not(effects.YouOwn())),
			),
		},
	})
}

// carrionLocustLoseLife is "if it was a creature card, that player loses 1
// life": the exiled card's owner. An on-disk key: never renamed, never reused.
var carrionLocustLoseLife = effects.CardTail("carrion-locust/lose-life", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.OneWentThisWay(o) || o.Tail.N == 0 {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(o.IDs[0])
	if !ok {
		return nil
	}
	return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), c.Owner, -1)
})
