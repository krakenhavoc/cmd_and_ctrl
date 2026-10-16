package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// loyalty_counter_trigger.go — #2797: "Whenever you put one or more
// loyalty counters on a planeswalker" (Inspired Tethermage).
//
// ONE TRIGGER PER PLACEMENT, NOT PER COUNTER. The engine emits one
// EventCounterPlaced per KIND per permanent per placement, carrying the
// total after the change, and "one or more" (CR 603.2c) is exactly that
// event: a +2 loyalty ability, a Doubling Season that turns one counter
// into two, and a single "put three loyalty counters" are each ONE event
// and so one trigger. A loyalty counter is one kind, so there is never a
// second event from the same placement to merge, which is why this needs
// no OncePerBatch.
//
// What it must NOT collapse is two different planeswalkers: "put a
// loyalty counter on each planeswalker you control" is one placement per
// permanent, and a trigger keyed on "a planeswalker" (an object, CR 122.6)
// fires once for each. OncePerBatch would have merged them into a single
// trigger and shipped the card weaker than printed; the batch dedup that
// Aragorn's "one or more counters on THIS permanent" needs is the wrong
// tool for "on a planeswalker".
//
// WHO PUT THEM is the event's Actor. A loyalty cost names its placer
// (the activating player, CR 606.4), and since #3942 so does every other
// placement: the controller of the resolving spell or ability, or the
// planeswalker's controller for the loyalty it enters with (CR 122.6a).
//
// A removal (a −N cost, damage) and a placement of nothing are not
// placements. A planeswalker entering with its loyalty counters IS: CR 122.6
// says "put" covers an object given counters as it enters.

// YouPutLoyaltyCountersOnAPlaneswalker is the trigger condition: ev is a
// placement of loyalty counters on a planeswalker on the battlefield,
// by the source's controller.
func YouPutLoyaltyCountersOnAPlaneswalker(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil || ev.Kind != game.EventCounterPlaced || ev.Label != game.CounterLoyalty {
		return false
	}
	target, ok := g.LookupCardForEffect(ev.Target)
	if !ok || !target.IsPlaneswalker() || !onBattlefield(g, ev.Target) {
		return false
	}
	_, placed := b12CountersPlacedBy(ev, source.Controller, g, true)
	return placed
}

// WheneverYouPutLoyaltyCountersOnAPlaneswalker is the printed ability.
// The trigger goes on the stack above whatever put the counters (CR 603.3b)
// and resolves first.
func WheneverYouPutLoyaltyCountersOnAPlaneswalker(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventCounterPlaced, YouPutLoyaltyCountersOnAPlaneswalker, label, effect)
}
