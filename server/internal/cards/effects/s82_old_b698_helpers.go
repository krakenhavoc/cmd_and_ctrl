package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b698_helpers.go — Mercadian Masques, "+1/+1 and other counters" (S82
// OLD batch 698, #5619). Append-only.

// S82OldB698MinusCounterOnBlocker is "put a -1/-1 counter on that creature" for
// a "becomes blocked by a creature" trigger (CR 509.3d: one per blocker): the
// counter goes on that blocker only if it is still the same object on the
// battlefield as the ability resolves.
func S82OldB698MinusCounterOnBlocker(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || item.Trigger.Object == nil {
		return nil
	}
	o := item.Trigger.Object
	if _, ok := g.PermanentForEffect(game.ObjectRef{ID: o.ID, Epoch: o.Epoch}); !ok {
		return nil
	}
	return AddCounter{Target: o.ID, Kind: game.CounterMinusOne, N: 1}.Apply(NewContext(g, item))
}

// S82OldB698PutCreatureOfManaValueXFromHand is "You may put a creature card
// with mana value X from your hand onto the battlefield", X the counters the
// cost removed (Mercadian Lift). A creature card with {X} in its cost has mana
// value 0 in hand (CR 202.3e).
func S82OldB698PutCreatureOfManaValueXFromHand(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		x := ctx.CountersRemoved()
		return PutFromHandOntoBattlefield{
			Match:    And(Creature(), ManaValueIs(x)),
			Optional: true,
			Label:    name + " — you may put a creature card with mana value X from your hand onto the battlefield",
		}.Apply(ctx)
	}
}

// S82OldB698ColorlessForXPlusOne is the mint of "Remove X <kind> counters:
// Add an amount of {C} equal to X plus one" (Kyren Toy): X is the count the
// cost removed.
func S82OldB698ColorlessForXPlusOne(_ *game.Game, _, _ uuid.UUID, paid game.PaidCost) game.ManaOutput {
	return game.ManaOutput{Amount: game.ManaAmount{Options: []string{"C"}, N: 1 + paid.CountersRemoved}}
}

// S82OldB698YoureDealtDamageMayPutCounters is Blood Hound's first ability:
// "Whenever you're dealt damage, you may put that many +1/+1 counters on this
// creature." The amount is what was applied (CR 119.3), read off the event; one
// trigger and one question for each damage event.
func S82OldB698YoureDealtDamageMayPutCounters(name string) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventDealDamage},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			_, ok := B43YouWereDealtDamage(ev, source, g)
			return ok
		},
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
			return game.NewTriggeredItem(source, name+" — you may put that many +1/+1 counters on it")
		},
		Effect: func(g *game.Game, item *game.StackItem) error {
			n := item.Trigger.Event.Amount
			if n <= 0 {
				return nil
			}
			return MayChoice{
				Player:   item.Controller,
				Question: name + " — put that many +1/+1 counters on it?",
				YesLabel: "Put them",
				NoLabel:  "Decline",
				Then:     s82OldB698MayPutCounters,
				Word:     strconv.Itoa(n),
			}.Apply(NewContext(g, item))
		},
	}
}

// s82OldB698MayPutCounters is the "yes": that many +1/+1 counters on the
// source, if it is still on the battlefield. An on-disk key: never renamed,
// never reused.
var s82OldB698MayPutCounters = MayYes("confirm/s82-old-b698-put-counters-on-this", func(ctx *Context, r game.ConfirmAnswer) error {
	return CounterOnThis{Kind: game.CounterPlusOne, N: WordInt(r)}.Apply(ctx)
})

// S82OldB698RemoveAllPlusOneCountersFromThis is "remove all +1/+1 counters from
// this creature".
func S82OldB698RemoveAllPlusOneCountersFromThis(g *game.Game, item *game.StackItem) error {
	c, ok := NewContext(g, item).SourcePermanent()
	if !ok {
		return nil
	}
	n := c.Counters[game.CounterPlusOne]
	if n <= 0 {
		return nil
	}
	return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: -n}.Apply(NewContext(g, item))
}

// S82OldB698LeyLineAsk is Ley Line's body: the upkeep player, not the
// controller, is asked whether to put the counter on the creature they chose.
func S82OldB698LeyLineAsk(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	who := item.Trigger.Event.Actor
	if who == uuid.Nil || g.PlayerByIDForEffect(who) == nil {
		return nil
	}
	return MayChoice{
		Player:   who,
		Question: "Ley Line — put a +1/+1 counter on the creature you chose?",
		Then:     s82OldB698LeyLinePut,
	}.Apply(NewContext(g, item))
}

// s82OldB698LeyLinePut is the upkeep player's "yes": a +1/+1 counter on the
// target, if it is still legal (CR 608.2b). An on-disk key.
var s82OldB698LeyLinePut = MayYes("confirm/s82-old-b698-ley-line-put", func(ctx *Context, _ game.ConfirmAnswer) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
		}
	}
	return nil
})
