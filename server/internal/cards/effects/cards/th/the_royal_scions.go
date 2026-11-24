package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Royal Scions — Legendary Planeswalker — Will Rowan {1}{U}{R}, loyalty 5:
//
//	"+1: Draw a card, then discard a card.
//	 +1: Target creature gets +2/+0 and gains first strike and trample until
//	 end of turn.
//	 −8: Draw four cards. When you do, The Royal Scions deals damage to any
//	 target equal to the number of cards in your hand."
//
// −8's "when you do" is a reflexive trigger (CR 603.12) that targets as it goes
// on the stack and counts the hand as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4ecde8d1-e4ec-4bd9-8a80-49885b032557",
		Name:            "The Royal Scions",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Draw a card, then discard a card.",
				Cost:    effects.LoyaltyCost(1),
				Purpose: game.Purpose{Draws: 1, Discards: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.DrawThenDiscard(g, item.Controller, item.SourceCardID, 1, 1, "The Royal Scions — draw a card, then discard a card")
				},
			},
			{
				Label:   "+1: Target creature gets +2/+0 and gains first strike and trample until end of turn.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetCreature("target creature"),
				Effect: effects.PumpAndGrantFirstTarget(2, 0, "The Royal Scions — +2/+0, first strike and trample",
					"first strike", "trample"),
			},
			{
				Label:   "−8: Draw four cards. When you do, The Royal Scions deals damage to any target equal to the number of cards in your hand.",
				Cost:    effects.LoyaltyCost(-8),
				Purpose: game.Purpose{Draws: 4},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.DrawCards{N: 4}).Apply(ctx); err != nil {
						return err
					}
					return effects.WhenYouDo("The Royal Scions — damage to any target equal to the cards in your hand", theRoyalScionsDamageBody).Apply(ctx)
				},
			},
		},
	})
}

// theRoyalScionsDamageBody is the reflexive trigger. An on-disk key: never
// renamed, never reused.
var theRoyalScionsDamageBody = game.ReflexiveBody("the-royal-scions/damage", effects.SimpleBody(
	func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		p := g.PlayerByIDForEffect(item.Controller)
		if p == nil || p.Hand == nil {
			return nil
		}
		n := p.Hand.Size()
		for _, t := range ctx.LegalTargets() {
			return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}.Apply(ctx)
		}
		return nil
	}), effects.ConstTargets(effects.TargetAny))
