package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reality Smasher — Creature — Eldrazi {4}{C}, 5/5 (Secret Lair Drop, #5462):
//
//	"Trample, haste
//	 Whenever this creature becomes the target of a spell an opponent
//	 controls, counter that spell unless its controller discards a card."
//
// Bonecrusher Giant's spell-only watch (SelfTargetedByASpell: an ability that
// targets it does nothing), narrowed to a spell an opponent controls. The
// trigger goes on the stack above the spell. Its controller is asked to discard
// a card, Argentum Masticore's keyed discard prompt, which holds the table
// while it is open so the spell can't resolve out from under it; a controller
// who discards nothing, or has no card to discard, has the spell countered
// (CounterTarget; a spell that can't be countered stays, and one already gone
// is not an error).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "17c651c7-da4e-45e9-9a64-fad3e873123d",
		Name:            "Reality Smasher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, realitySmasherTargeted,
				"Reality Smasher — counter that spell unless its controller discards a card",
				realitySmasherAsk),
		},
	})
}

// realitySmasherTargeted is "becomes the target of a spell an opponent
// controls".
func realitySmasherTargeted(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
	return effects.SelfTargetedByASpell(ev, source, ch, g) && ev.Actor != source.Controller
}

// realitySmasherAsk asks the spell's controller for the discard, or counters
// the spell at once when there is nothing to discard.
func realitySmasherAsk(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ev := item.Trigger.Event
	ctx := effects.NewContext(g, item)
	p := g.PlayerByIDForEffect(ev.Actor)
	if p == nil || p.Eliminated || p.Hand == nil || p.Hand.Size() == 0 {
		return effects.CounterTarget{StackID: ev.StackItemID}.Apply(ctx)
	}
	return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
		Player:   ev.Actor,
		Source:   item.SourceCardID,
		N:        1,
		UpTo:     true,
		Question: "Reality Smasher — discard a card, or the spell is countered?",
		Asked:    game.NewAskedItem(item),
	}, realitySmasherDiscarded, nil)
}

// realitySmasherDiscarded counters the spell when nothing was discarded. Keyed
// (#4253); an on-disk key: never renamed, never reused.
var realitySmasherDiscarded = game.RegisterDiscardThen("choose-cards/reality-smasher-discard", func(g *game.Game, r game.DiscardsDone) error {
	if r.Item == nil || r.Item.Trigger == nil || len(r.Discarded.Cards()) > 0 {
		return nil
	}
	return effects.CounterTarget{StackID: r.Item.Trigger.Event.StackItemID}.Apply(effects.NewContext(g, r.Item))
})
