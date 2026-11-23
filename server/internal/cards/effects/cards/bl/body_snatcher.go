package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Body Snatcher — Creature — Phyrexian Minion {2}{B}{B}, 2/2 (Dominaria
// Remastered):
//
//	"When this creature enters, exile it unless you discard a creature
//	 card.
//	 When this creature dies, exile it and return target creature card
//	 from your graveyard to the battlefield."
//
// The enter trigger offers a discard of up to one card with a set rule
// that the card must be a creature card (Argentum Masticore's "unless you
// discard", Teferi Akosa's rule): choosing none, or holding no creature
// card, exiles the Snatcher, but only while it is still the object that
// entered (CR 400.7). The dies trigger targets a creature card in your
// graveyard (it may be the Snatcher itself, which is there when the
// trigger goes on the stack). It exiles the Snatcher if it is still the
// card that died in the graveyard (Moldgraf Monstrosity), and then, whether
// or not that worked, returns the target to the battlefield if it is still
// in the graveyard (CR 608.2b). A Snatcher that exiled itself is no longer
// there to come back.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e04688e0-672b-407f-a2c9-fd7292926483",
		Name:         "Body Snatcher",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Body Snatcher — exile it unless you discard a creature card", bodySnatcherEnters),
			effects.Targeting(
				effects.WhenThisDies("Body Snatcher — exile it and return target creature card from your graveyard to the battlefield",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						if !effects.DiedCardStillInGraveyard(ctx, item.Trigger.Event.CardID) {
							bodySnatcherReturn(ctx)
							return nil
						}
						return effects.ExileTarget{
							Target:   item.Trigger.Event.CardID,
							ThenTail: ctx.ItemTail(bodySnatcherExiled),
						}.Apply(ctx)
					}),
				effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn()),
			),
		},
	})
}

// bodySnatcherExiled is the clause after the exile: the return, whether or
// not the exile happened. An on-disk key.
var bodySnatcherExiled = effects.CardTail("body-snatcher/exiled", func(ctx *effects.Context, _ game.TailOutcome) error {
	bodySnatcherReturn(ctx)
	return nil
})

// bodySnatcherReturn returns the targeted creature card from your graveyard
// to the battlefield under your control.
func bodySnatcherReturn(ctx *effects.Context) {
	_, _ = effects.ReanimateSingleTarget(ctx, ctx.Controller())
}

// bodySnatcherEnters asks for the discard, or exiles the Snatcher at once
// when there is no creature card to discard.
func bodySnatcherEnters(g *game.Game, item *game.StackItem) error {
	if !effects.SourceIsStillThisPermanent(g, item) {
		return nil
	}
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil || p.Eliminated {
		return nil
	}
	hasCreature := false
	if p.Hand != nil {
		for _, c := range p.Hand.Cards {
			if c.IsCreature() {
				hasCreature = true
			}
		}
	}
	if !hasCreature {
		return effects.ExileTarget{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
	}
	return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
		Player:   item.Controller,
		Source:   item.SourceCardID,
		N:        1,
		UpTo:     true,
		Question: "Body Snatcher — discard a creature card, or exile it?",
		Asked:    game.NewAskedItem(item),
	}, bodySnatcherDiscarded, nil)
}

// bodySnatcherDiscarded is the discard answered. The set rule is "a creature
// card": a pick of anything else is refused, and no pick at all is the
// decline, which exiles the Snatcher if it is still the same object. Keyed
// (#4253); an on-disk key: never renamed, never reused.
var bodySnatcherDiscarded = game.RegisterDiscardThenRule("choose-cards/body-snatcher-discard",
	func(picked []game.Card) bool { return len(picked) == 1 && picked[0].IsCreature() },
	func(g *game.Game, r game.DiscardsDone) error {
		if r.Item == nil || len(r.Discarded.Cards()) > 0 {
			return nil
		}
		if !effects.SourceIsStillThisPermanent(g, r.Item) {
			return nil
		}
		return effects.ExileTarget{Target: r.Item.SourceCardID}.Apply(effects.NewContext(g, r.Item))
	})
