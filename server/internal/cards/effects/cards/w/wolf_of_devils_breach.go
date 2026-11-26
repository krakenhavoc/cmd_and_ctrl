package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wolf of Devil's Breach — Creature — Elemental Wolf {3}{R}{R}, 5/5 (Shadows
// over Innistrad, S82 OLD):
//
//	"Whenever this creature attacks, you may pay {1}{R} and discard a card. If
//	 you do, this creature deals damage to target creature or planeswalker
//	 equal to the discarded card's mana value."
//
// The target is chosen as the trigger goes on the stack, and the payment is
// asked as it resolves (Banon's "pay {1} and discard a card"): the mana first,
// then a discard prompt with a floor of one, so both halves are owed once the
// player says yes. With an empty hand nothing is offered. The damage is the
// Wolf's, as an object, and equals the mana value of the card that was really
// discarded, read in the graveyard (Argentum Masticore's reading).
//
// No simplification.
func init() {
	const label = "Wolf of Devil's Breach — you may pay {1}{R} and discard a card: damage to target creature or planeswalker equal to its mana value"
	t := effects.WheneverThisAttacks(label, func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		if effects.B14HandSize(g, item.Controller) == 0 {
			return nil
		}
		return effects.MayPay{
			Chooser:  item.Controller,
			Cost:     "{1}{R}",
			Question: "Wolf of Devil's Breach — pay {1}{R} and discard a card to deal damage to the target?",
			Then:     wolfOfDevilsBreachPaid,
		}.Apply(ctx)
	})
	t.Targets = effects.Harmful(effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())))
	effects.Register(effects.Spec{
		OracleID:     "453e72bf-a682-4ec8-9cdd-799f2acb3862",
		Name:         "Wolf of Devil's Breach",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}

// wolfOfDevilsBreachPaid is the paid half: the discard, keyed (#4264).
var wolfOfDevilsBreachPaid = effects.PayUnlessBranches("pay-unless/wolf-of-devils-breach-discard",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
			Player:   ctx.Controller(),
			Source:   ctx.Source(),
			N:        1,
			Question: "Wolf of Devil's Breach — discard a card",
			Asked:    game.NewAskedItem(ctx.Item),
		}, wolfOfDevilsBreachDiscarded, nil)
	}, nil)

// wolfOfDevilsBreachDiscarded is "if you do": the Wolf deals damage equal to
// the discarded card's mana value to the trigger's target. An on-disk key.
var wolfOfDevilsBreachDiscarded = game.RegisterDiscardThen("choose-cards/wolf-of-devils-breach-damage", func(g *game.Game, r game.DiscardsDone) error {
	discarded := r.Discarded.Cards()
	if r.Item == nil || len(discarded) == 0 {
		return nil
	}
	card, ok := g.LookupCardForEffect(discarded[0])
	if !ok {
		return nil
	}
	ctx := effects.NewContext(g, r.Item)
	target, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	return effects.DealDamage{SourceObject: &ref, Target: target, Amount: card.ManaValue()}.Apply(ctx)
})
