package game

import "github.com/google/uuid"

// event_batch.go — what "one or more" counts (#829, CR 603.2c).
//
// CR 603.2c: "An ability triggers only once each time its trigger
// event occurs." A "whenever ONE OR MORE creatures you control leave
// the battlefield" ability watches an event the engine emits once per
// creature, so it needs to know which of those events happened AT THE
// SAME TIME. That set is an event BATCH, and a batch is the only unit
// OncePerBatch collapses: the ability fires once for a batch, and
// again for the next one.
//
// Before this, "the batch" was modelled as "nothing of this ability is
// in flight" — a scan of PendingTriggers, StackMeta and the open
// trigger prompts. Two of those three are bounded by the next
// priority boundary and were a fair proxy; the stack is not. An item
// on the stack outlives the batch that put it there, so a SECOND,
// separate batch arriving while the first batch's trigger waited to
// resolve was swallowed as if it were part of the first. Dour
// Port-Mage drew one card for two bounces (#829, and #587's original
// description of the same failure).
//
// # The boundary
//
// A batch is every event the engine emits between two points where
// PLAY MOVES ON — and there are exactly two such points:
//
//   - a stack item begins to resolve (resolveTopOfStackLocked), and
//   - the turn cursor enters a new step (advanceCursorLocked), plus
//     the one step the RULES split in two and the cursor does not:
//     the first-strike and regular combat damage steps (CR 510.4),
//     which resolveCombatDamageLocked runs back to back inside the
//     cursor's single combat_damage step (#784).
//
// Nothing else opens a batch. That is deliberate, and the two
// consequences are the ones the rules want:
//
//   - Everything one resolution emits is one batch. A Cyclonic Rift
//     that bounces four creatures is one occurrence, and Dour
//     Port-Mage draws one card.
//   - Everything one turn-based action emits is one batch, however
//     many engine calls the sandbox splits it across. Declaring
//     attackers one creature at a time (DeclareAttacker, the verb the
//     client still sends per click) stays ONE declaration, because no
//     step change and no resolution happens between the clicks — so
//     Adeline makes one batch of Humans for a three-creature attack
//     whether the seat used the per-creature verb or the bulk one.
//
// The known gap is the mirror image of that second point: two
// SANDBOX-MANUAL mutations in the same priority window (two
// move_card bounces in a row, with nothing resolving in between) share
// a batch and fire a "one or more" trigger once. That is what the
// old in-flight check did too, so it is not a regression, and manual
// zone shoving has no rules occurrence to count in the first place.
//
// # The key, and its second dimension (#784)
//
// A batch is WHEN; the key is WHAT. CR 603.2c's other half — "it can
// trigger repeatedly if one event contains multiple occurrences" —
// is what a clause naming an object asks for: "whenever one or more
// creatures you control deal combat damage to A PLAYER" is one
// trigger per player dealt damage in that damage step, and "whenever
// you attack A PLAYER" one per player attacked in that declaration.
// TriggeredAbility.BatchKey reads that object off the event and
// oncePerBatchKeyLocked appends it to the ability's Key, so the guard
// reads "(source, key, object) once per batch". An ability whose
// clause names no object leaves BatchKey nil and is unchanged. One
// key with two dimensions, one guard, still no second dedupe path —
// which is what let #784 delete the last hand-rolled one
// (TriggerInFlightForEffect, Breena and Nature's Will).

// beginEventBatchLocked opens a new event batch. Every Event emitted
// from here until the next call is stamped with it, and a
// OncePerBatch ability that already fired for an earlier batch fires
// again for this one.
//
// Called from exactly two places — resolveTopOfStackLocked and
// advanceCursorLocked — which together are "play moved on". See the
// file comment for why those two and nothing else.
//
// Caller must hold g.mu in write mode.
func (g *Game) beginEventBatchLocked() {
	g.eventBatch++
}

// currentEventBatchLocked is the batch EmitEvent stamps. The counter
// starts at zero on a freshly constructed Game and the first event of
// a game would then carry the un-stamped sentinel, so the first read
// opens batch 1. Caller must hold g.mu in write mode.
func (g *Game) currentEventBatchLocked() uint64 {
	if g.eventBatch == 0 {
		g.eventBatch = 1
	}
	return g.eventBatch
}

// oncePerBatchKeyLocked is the ability key the guard below is keyed
// on: TriggeredAbility.Key, plus the dimension BatchKey reads off the
// event when the printed clause quantifies over an object (#784).
//
// One key, two dimensions — not a second dedupe path. A clause that
// names no object (Dour Port-Mage's "one or more creatures leave the
// battlefield") has a nil BatchKey and keys on the label alone; one
// that names a player ("deal combat damage to A PLAYER", "attack A
// PLAYER") appends that player, so the batch collapses per player
// instead of per step (CR 603.2c).
//
// Caller must hold g.mu in write mode.
func oncePerBatchKeyLocked(t TriggeredAbility, ev Event, source *Card, g *Game) string {
	if t.BatchKey == nil {
		return t.Key
	}
	return t.Key + "|" + t.BatchKey(ev, source, g)
}

// oncePerBatchAllowsLocked is the CR 603.2c guard behind
// TriggeredAbility.OncePerBatch: it reports whether the ability named
// by (source, key) may still fire for the batch `batch`, and records
// it as fired when it may.
//
// Test-and-set in one call, because the two halves must not drift:
// every caller that asks is about to dispatch.
//
// `key` is oncePerBatchKeyLocked's output — the ability's Key with
// its per-event dimension appended when it has one. An empty key
// means "any trigger from this source", matching
// TriggeredAbility.Key's own contract; TallyKey handles it the same
// way TurnTally.Triggered does.
//
// Caller must hold g.mu in write mode.
func (g *Game) oncePerBatchAllowsLocked(batch uint64, source uuid.UUID, key string) bool {
	k := TallyKey(source, key)
	if fired, ok := g.oncePerBatchFired[k]; ok && fired == batch {
		return false
	}
	if g.oncePerBatchFired == nil {
		g.oncePerBatchFired = map[string]uint64{}
	}
	g.oncePerBatchFired[k] = batch
	return true
}
