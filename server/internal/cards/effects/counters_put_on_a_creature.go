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
// Who put them is b12CountersPlacedByYou's attribution rule; the
// creature is judged as it is on the battlefield when the event is
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
