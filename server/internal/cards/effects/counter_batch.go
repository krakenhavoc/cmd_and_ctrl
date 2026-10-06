package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// counter_batch.go — "whenever you put one or more counters on ~" with
// the KINDS put together (#2150, CR 122.6 and 603.2c; Aragorn, Company
// Leader).
//
// The engine emits one EventCounterPlaced per KIND per permanent, so a
// placement of two kinds at once is two events. An ability that says
// "one or more counters" is one occurrence for the whole placement
// (CR 603.2c), and one that copies "one of each of those kinds" has to
// know every kind in it. Read naively the pair is two triggers that
// could each pick a different target: stronger than printed.
//
// The two halves are the ones the engine already has:
//
//   - WHEN is the event batch (event_batch.go): every event emitted
//     between two points where play moves on is one occurrence, and
//     OncePerBatch fires a trigger once per batch. The ability is
//     declared OncePerBatch, so the first qualifying kind of a batch
//     triggers it and the rest are declined.
//   - WHICH is read back as the trigger RESOLVES: every
//     EventCounterPlaced of the triggering event's batch on the same
//     permanent, credited to the controller by the same attribution rule
//     as All Will Be One (b12CountersPlacedBy). By then the batch is
//     complete, because the trigger resolves after the resolution that
//     emitted its first event has finished.
//
// What the batch cannot tell apart is two separate placements inside
// one resolution ("put a +1/+1 counter on it. Then put a flying counter
// on it."): the engine batches the whole resolution, so they read as one
// occurrence. The error is toward FEWER triggers, never more.

// WheneverYouPutOneOrMoreCountersOnThis is "whenever you put one or
// more counters on this permanent": one trigger for the whole batch of
// counters placed on it at the same time, whatever their kinds. The
// effect reads which kinds with CountersYouPutInTheSameBatch.
func WheneverYouPutOneOrMoreCountersOnThis(label string, effect Effect) game.TriggeredAbility {
	return OncePerBatch(On(game.EventCounterPlaced, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if source == nil || ev.Target != source.InstanceID {
			return false
		}
		_, ok := b12CountersPlacedByYou(ev, source, g)
		return ok
	}, label, effect))
}

// CountersYouPutInTheSameBatch is the distinct kinds of counter that
// `you` put on ev's permanent in the event batch ev belongs to, in the
// order they were placed. ev is the triggering event, which a trigger's
// effect reads from item.Trigger.Event. A removal, a placement of zero
// and a placement by another player are not in it.
//
// Caller must hold g.mu.
func CountersYouPutInTheSameBatch(g *game.Game, ev game.Event, you uuid.UUID) []string {
	var kinds []string
	seen := map[string]bool{}
	for i := len(g.Events) - 1; i >= 0; i-- {
		e := g.Events[i]
		if e.Batch < ev.Batch {
			break
		}
		if e.Batch != ev.Batch || e.Kind != game.EventCounterPlaced || e.Target != ev.Target || seen[e.Label] {
			continue
		}
		if _, ok := b12CountersPlacedBy(e, you, g, false); !ok {
			continue
		}
		seen[e.Label] = true
		kinds = append(kinds, e.Label)
	}
	// Collected newest first; hand them back in the order placed.
	for l, r := 0, len(kinds)-1; l < r; l, r = l+1, r-1 {
		kinds[l], kinds[r] = kinds[r], kinds[l]
	}
	return kinds
}
