package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// counters_put_on_a_creature.go — "whenever you put one or more counters
// on a creature [you don't control]" (Generous Patron, Rikku, Resourceful
// Guardian; #3246).
//
// WheneverYouPutOneOrMoreCountersOnThis (counter_batch.go) is the same
// clause about one permanent. Here the permanent is any creature that
// `match` admits, and the occurrence is per CREATURE: Generous Patron's
// support 2 on two opponents' creatures is two placements and two
// triggers (CR 603.2c), while two kinds of counter put on one creature at
// once is still one. So the ability is OncePerBatch with the batch keyed
// on the permanent the counters went on.
//
// Who put them is the event's recorded placer (b12CountersPlacedByYou);
// the creature is judged as it is on the battlefield when the event is
// harvested.

// WheneverYouPutOneOrMoreCountersOnACreature fires once per creature per
// batch of counters you put on it, for a creature `match` admits (nil
// admits every creature). The effect reads the creature from
// item.Trigger.Event.Target.
func WheneverYouPutOneOrMoreCountersOnACreature(match func(g *game.Game, source *game.Card, c game.Card) bool, label string, effect Effect) game.TriggeredAbility {
	t := OncePerBatch(On(game.EventCounterPlaced, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if source == nil {
			return false
		}
		if _, ok := b12CountersPlacedByYou(ev, source, g); !ok {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.Target)
		if !ok || !c.IsCreature() {
			return false
		}
		return match == nil || match(g, source, c)
	}, label, effect))
	t.BatchKey = counterTargetBatchKey
	return t
}

// counterTargetBatchKey is the TriggeredAbility.BatchKey for a counter
// placement: the permanent the counters went on.
func counterTargetBatchKey(ev game.Event, _ *game.Card, _ *game.Game) string {
	return ev.Target.String()
}

// creatureYouDontControl is the match for "a creature you don't
// control".
func creatureYouDontControl(_ *game.Game, source *game.Card, c game.Card) bool {
	return c.Controller != source.Controller && c.Controller != uuid.Nil
}

// creatureYouControl is the match for "a creature you control".
func creatureYouControl(_ *game.Game, source *game.Card, c game.Card) bool {
	return c.Controller == source.Controller && c.Controller != uuid.Nil
}

// WheneverYouPutCountersOfKindOnACreature is
// WheneverYouPutOneOrMoreCountersOnACreature narrowed to one counter
// kind: "whenever you put one or more +1/+1 counters on a creature you
// control" (Terrasymbiosis, #3929). Once per creature per batch, like
// its sibling; a batch that put only other kinds on the creature does
// not trigger it. The effect reads "that many" with
// CountersOfKindYouPutInTheSameBatch.
//
// Who put them is countersPutBy: the event's recorded placer.
func WheneverYouPutCountersOfKindOnACreature(kind string, match func(g *game.Game, source *game.Card, c game.Card) bool, label string, effect Effect) game.TriggeredAbility {
	t := OncePerBatch(On(game.EventCounterPlaced, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if source == nil || ev.Label != kind {
			return false
		}
		if _, ok := countersPutBy(ev, source.Controller, g, true); !ok {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.Target)
		if !ok || !c.IsCreature() {
			return false
		}
		return match == nil || match(g, source, c)
	}, label, effect))
	t.BatchKey = counterTargetBatchKey
	return t
}

// CountersOfKindYouPutInTheSameBatch is how many counters of `kind`
// `you` put on ev's permanent in the event batch ev belongs to: the sum
// of each placement's delta. The engine's counter event carries the
// total after the change, so countersPutBy recovers each delta from the
// log. A removal and a placement by another player add nothing. ev is
// the triggering event (item.Trigger.Event), read as the trigger
// resolves, by when the batch is complete.
//
// Caller must hold g.mu.
func CountersOfKindYouPutInTheSameBatch(g *game.Game, ev game.Event, you uuid.UUID, kind string) int {
	n := 0
	for i := len(g.Events) - 1; i >= 0; i-- {
		e := g.Events[i]
		if e.Batch < ev.Batch {
			break
		}
		if e.Batch != ev.Batch || e.Kind != game.EventCounterPlaced || e.Target != ev.Target || e.Label != kind {
			continue
		}
		if delta, ok := countersPutBy(e, you, g, false); ok {
			n += delta
		}
	}
	return n
}

// countersPutBy is how many counters ev placed and whether `you` put
// them: b12CountersPlacedBy, which reads the placer every placement
// names on the event (Event.Actor, #3942).
func countersPutBy(ev game.Event, you uuid.UUID, g *game.Game, requireBattlefield bool) (int, bool) {
	return b12CountersPlacedBy(ev, you, g, requireBattlefield)
}
