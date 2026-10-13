package game

import (
	"strconv"

	"github.com/google/uuid"
)

// batch_total.go — "that much" and "that many" over simultaneous
// events (#3172 slice cav-d, ADR 0108 amendment 2026-10-13, CR 120.3,
// CR 510.2, CR 603.2c).
//
// The engine opens one EventDealDamage per (source, recipient), one
// EventDiscardCard per card and one EventETB per permanent. A printed
// "whenever ~ is dealt damage, … that much" or "whenever you discard
// one or more cards, … that many" is ONE trigger reading the TOTAL of
// what happened at the same time. Reading the events one at a time
// triggered once per source, so a Dragon blocked by two creatures
// reflected twice, with two target picks, where the printed card
// reflects once for the sum — and a per-instance damage booster
// (Torbran) counted twice.
//
// It is the second reading of AtBatchEnd (life_batch.go): the ability
// is staged when its first matching event arrives and settled once the
// batch is over, and the settled event is the FIRST matching event with
// its Amount rewritten to the total. A card reads
// `item.Trigger.Event.Amount`; the Build it already had is unchanged.
//
// # The two totals
//
//   - BatchTotalDamage: the sum of the damage dealt to the staged
//     event's Target in the staged event's damage instance — the
//     events one combat damage step opened (CR 510.2), or one damage
//     instruction (ADR 0108 owner decision 4). Each damaged object is
//     its own group, so Blasphemous Act is one trigger per creature,
//     and two separate instructions on one creature in the same
//     resolution are two groups, because they are two events.
//   - BatchTotalEvents: the number of events the ability matches in
//     the batch. One group for the whole batch, the OncePerBatch unit
//     (CR 603.2c), so two instructions of one resolution count as one
//     occurrence, which is the same approximation every "one or more"
//     card already makes.
//
// Both re-run the ability's own AppliesTo over the logged events, so a
// filter such as "artifacts you control" or "cards you discard" applies
// to the sum exactly as it applied to the first event.
//
// The total never reports less than the first event's own amount, so a
// log that lost events (or a restored game that dropped the transient
// DamageInstance stamp and groups on the batch alone) errs weaker than
// printed, never stronger.

// BatchTotal names the sum an AtBatchEnd ability settles with.
type BatchTotal string

const (
	// BatchTotalNone is the life-loss reading of AtBatchEnd.
	BatchTotalNone BatchTotal = ""
	// BatchTotalDamage sums Amount over the EventDealDamage events one
	// damage instance dealt to one object.
	BatchTotalDamage BatchTotal = "damage"
	// BatchTotalEvents counts the batch's matching events.
	BatchTotalEvents BatchTotal = "events"
)

// group is the staging key's second dimension for an event: one
// staged entry per distinct value, so two damaged creatures are two
// triggers and one creature damaged twice by one step is one.
func (c *BatchEndCondition) group(ev Event) string {
	if c == nil || c.Total != BatchTotalDamage {
		return ""
	}
	return ev.Target.String() + "|" + strconv.FormatUint(uint64(ev.DamageInstance), 10)
}

// batchTotalLocked is the settled total for one staged match: the
// events the ability watches, in the staged event's batch (and, for
// damage, on its object and instance), that the ability's AppliesTo
// still accepts.
//
// Caller must hold g.mu.
func (g *Game) batchTotalLocked(src *Card, lki Characteristic, t TriggeredAbility, first Event) int {
	total := 0
	// Events are appended in batch order, so walk back from the end and
	// stop once the batch is behind us.
	for i := len(g.Events) - 1; i >= 0; i-- {
		ev := g.Events[i]
		if ev.Batch < first.Batch {
			break
		}
		if ev.Batch != first.Batch || !triggerWatches(t.Watches, ev.Kind) {
			continue
		}
		weight := 1
		if t.AtBatchEnd.Total == BatchTotalDamage {
			if ev.Target == uuid.Nil || ev.Target != first.Target || ev.DamageInstance != first.DamageInstance {
				continue
			}
			weight = ev.Amount
		}
		if weight <= 0 {
			continue
		}
		if t.AppliesTo != nil && !t.AppliesTo(ev, src, lki, g) {
			continue
		}
		total += weight
	}
	if t.AtBatchEnd.Total == BatchTotalEvents && total < 1 {
		total = 1
	}
	if t.AtBatchEnd.Total == BatchTotalDamage && total < first.Amount {
		total = first.Amount
	}
	return total
}
