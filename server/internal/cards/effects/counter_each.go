package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// counter_each.go — "whenever a [kind] counter is put on ~" (#1841).
//
// WheneverACounterIsPutOnThis is the one-trigger-PER-COUNTER reading
// (CR 603.2c, Fathom Mage's rulings), the opposite of
// WheneverYouPutOneOrMoreCountersOnThis in counter_batch.go. The
// engine emits one EventCounterPlaced per placement, after any
// replacement has settled how many counters it was (Doubling Season,
// Hardened Scales), and game.TriggeredAbility.PerCounter turns the
// number of counters that event put into that many occurrences: each
// its own stack object and each its own "you may". Counters a
// permanent enters with are placed after it arrives, so its own
// ability sees them (CR 122.6). It is NOT OncePerBatch.
func WheneverACounterIsPutOnThis(kind, label string, effect Effect) game.TriggeredAbility {
	t := On(game.EventCounterPlaced, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return source != nil && ev.Target == source.InstanceID && ev.Label == kind
	}, label, effect)
	t.PerCounter = kind
	return t
}

// WheneverACounterIsPutOnACreature is "whenever a [kind] counter is
// put on a creature" — any creature, whoever controls it or put the
// counter there, once per counter (Flourishing Defenses).
func WheneverACounterIsPutOnACreature(kind, label string, effect Effect) game.TriggeredAbility {
	t := On(game.EventCounterPlaced, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
		if ev.Label != kind {
			return false
		}
		z := g.FindCardZoneForEffect(ev.Target)
		if z == nil || z.Kind != game.ZoneBattlefield {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.Target)
		return ok && c.IsCreature()
	}, label, effect)
	t.PerCounter = kind
	return t
}

// WheneverACounterIsRemovedFromThis is "whenever a [kind] counter is
// removed from ~" — the removal twin of WheneverACounterIsPutOnThis
// (#2466), once per counter (CR 603.2c, Protean Hydra's ruling). Any
// cause counts: a cost, an effect, prevented damage. Counters that
// vanish because the permanent left the battlefield do not (CR 122.2).
// The `zones` are where the ability works — none for the battlefield,
// ZoneExile for a suspended card's "while it's exiled".
func WheneverACounterIsRemovedFromThis(kind, label string, effect Effect, zones ...game.ZoneKind) game.TriggeredAbility {
	t := On(game.EventCounterPlaced, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return source != nil && ev.Target == source.InstanceID && ev.Label == kind
	}, label, effect)
	t.PerCounterRemoved = kind
	t.Zones = zones
	return t
}
