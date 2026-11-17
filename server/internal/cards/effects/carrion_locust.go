package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "04e5ecda-9646-4bc3-addf-cca471553fbe",
		Name:            "Carrion Locust",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Carrion Locust — exile target card from an opponent's graveyard; if it was a creature card, that player loses 1 life",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							return ExileThenIfItWas{Target: t.ID, Was: WasCreatureCard, Keyed: carrionLocustLoseLife}.Apply(ctx)
						}
						return nil
					}),
				TargetCardInGraveyard("target card from an opponent's graveyard", Not(YouOwn())),
			),
		},
	})
}

// carrionLocustLoseLife is "if it was a creature card, that player loses 1
// life": the exiled card's owner. An on-disk key: never renamed, never reused.
var carrionLocustLoseLife = CardTail("carrion-locust/lose-life", func(ctx *Context, o game.TailOutcome) error {
	if !oneWentThisWay(o) || o.Tail.N == 0 {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(o.IDs[0])
	if !ok {
		return nil
	}
	return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), c.Owner, -1)
})
