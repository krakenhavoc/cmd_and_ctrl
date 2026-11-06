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
// PLAY MOVES ON — and there are exactly four such points:
//
//   - a stack item begins to resolve (resolveTopOfStackLocked), and
//   - the turn cursor enters a new step (advanceCursorLocked), plus
//     the one step the RULES begin again where the cursor does not
//     move: the extra cleanup step CR 514.3a asks for, which
//     repeatCleanupStepLocked opens (#661, cleanup.go), and
//   - a special action is taken (PerformSpecialAction, #1341). CR
//     116.2 makes each one — foretell, suspend, turning a face-down
//     permanent up — its own event with no stack and no priority
//     pass around it, so it is a third point where play moves on:
//     two foretells taken back to back, before anything resolves,
//     are two occurrences (CR 603.2c), not one, and
//   - a player casts a spell or plays a land as their own action
//     (castSpellLocked, #4943). CR 601.2a moves the card to the stack
//     as the cast begins, and CR 116.2a makes a land play a special
//     action that reaches the engine through the cast verb rather
//     than PerformSpecialAction, so each is its own event for the
//     same reason a foretell is: two flashback casts in one priority
//     window, or a land played from the graveyard and then a spell
//     cast from it, are two departures from the graveyard and two
//     triggers of Tormod, the Desecrator, not one. The batch opens
//     once the cast has passed its announce-time gates and any
//     question it parks on has been answered (a parked cast opens it
//     on the resume that makes it), so everything the cast emits —
//     its auto-tap, its cost payments (a delve's exiles, a discard),
//     the move to the stack and EventCast — is one occurrence. A cast
//     or land play made INSIDE a resolution (CR 608.2g: a copy a
//     resolution offers, Word of Command's play) is part of that
//     resolution's occurrence and opens nothing; see
//     beginActionEventBatchLocked.
//
// An activated ability (CR 602) does not open one yet: two
// activations in one window whose costs exile cards from a graveyard
// share a batch, as two casts did before #4943.
//
// The step half used to carry a second exception. The first-strike
// and regular combat damage steps are TWO combat damage steps
// (CR 510.4), so a first-striker and a regular attacker connecting
// with the same player are two occurrences — and #784 had to open a
// batch by hand between the two passes because both ran inside the
// cursor's single combat_damage step. Since #717 the cursor really
// does enter two steps, so the ordinary boundary produces the two
// batches and that hand-rolled one is gone with the step it worked
// around.
//
// Nothing else opens a batch. That is deliberate, and the two
// consequences below are unchanged by the special-action point added
// in #1341 and the cast point added in #4943 — neither is ever inside
// a resolution or a turn-based action, so each only ever adds a
// boundary, never removes one:
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
// Called from resolveTopOfStackLocked, advanceCursorLocked,
// repeatCleanupStepLocked, PerformSpecialAction and (through
// beginActionEventBatchLocked) castSpellLocked — a resolution
// beginning, the cursor entering a step, the one step that begins
// again without the cursor moving (CR 514.3a), a special action
// being taken (CR 116.2, #1341), and a spell cast or a land played as
// a player's own action (CR 601.2a, 116.2a, #4943). Together they are
// "play moved on". See the file comment for why those and nothing
// else.
//
// It is also where the RESOLVING-ITEM slot is cleared (#920,
// resolving_item.go). The two are the same question asked twice: a
// resolving spell's metadata has to outlive its own resolution
// function, because the copy decision it opens is a prompt answered
// afterwards — and it has to stop being reachable the moment play
// moves on, which is precisely this boundary. A resolution-time prompt
// blocks the table, so an open copy decision cannot be crossed by
// either half of it.
//
// Caller must hold g.mu in write mode.
func (g *Game) beginEventBatchLocked() {
	// ADR 0107 §6: the instance of damage the batch held is over, so
	// the follow-ups it owes run now (CR 615.5), if the priority
	// boundary has not already run them — in the batch they belong to.
	g.flushPreventionFollowUpsLocked()
	// #2183: an AtBatchEnd trigger is owed its answer from the batch
	// that is ending, before the counter moves on.
	g.settleBatchEndTriggersLocked()
	// #1729, CR 610.3: an "until" return whose event happened in the
	// batch that is ending is owed before play moves on.
	g.resolveUntilReturnsLocked()
	g.eventBatch++
	g.resolving = nil
	// ADR 0107 §6: a next-damage shield spent in the batch that just
	// ended has prevented its instance of damage (CR 615.8).
	g.dropSpentNextShieldsLocked()
	// ADR 0108 §7: so have the divided shields' instances, unless one is
	// still waiting on its prompt.
	g.dropShieldDivisionsLocked(0)
	g.dropRedirectOrdersLocked(0)
}

// beginActionEventBatchLocked opens the batch a cast or a land play is
// when a player makes it as their own action (#4943, CR 601.2a,
// 116.2a), and does nothing when it is made inside a resolution.
//
// A resolution is open (resolutionOpen) from the moment its item begins
// to resolve until the CR 704.3 boundary after it, which waits for every
// prompt the resolution queued. A cast made in that span is one the
// resolution makes, or lets a player make as its answer (CR 608.2g: a
// resolution's copy offer, a "you may cast it" prompt, Word of Command
// playing the controlled player's card), and its events belong to the
// resolution's batch: everything one resolution emits is one occurrence,
// and the resolving slot the batch boundary clears is still being read.
// resolutionDepth is the same answer for a resolution function still on
// the Go stack.
//
// Caller must hold g.mu in write mode.
func (g *Game) beginActionEventBatchLocked() {
	if g.resolutionOpen || g.resolutionDepth > 0 {
		return
	}
	g.beginEventBatchLocked()
}

// BeginEventBatchForTest opens a new event batch from outside the package:
// play has moved on, so the resolving slot is empty. A test that answers a
// keyed prompt in the same call as the resolution that asked it would
// otherwise answer inside the batch the item resolved in, which a real
// answer never does.
func (g *Game) BeginEventBatchForTest() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.beginEventBatchLocked()
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

// CurrentEventBatchForEffect is the event batch a resolving effect's
// events are stamped with: inside a resolution, the batch it opened, so
// "for each card drawn this way" can count the draws this resolution
// made off this turn's log (Struggle for Project Purity, #2809). A read;
// it opens batch 1 on a game that has never emitted an event, as
// EmitEvent would. Caller must hold g.mu in write mode.
func (g *Game) CurrentEventBatchForEffect() uint64 {
	return g.currentEventBatchLocked()
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
