package game

// spell_entry_counters_mark.go — #4686: "that creature enters with X
// additional +1/+1 counters on it", given to a spell that is ALREADY on
// the stack by an effect (Wildgrowth Archaic's cast trigger).
//
// THE READ SIDE IS NOT NEW. StackItem.PromisedCounters (#1852,
// next_spell_promise.go) is seeded onto the spell's entry event by
// applyCastEntryCountersLocked, before the CR 614 pipeline, so Doubling
// Season and Hardened Scales see the counters and the permanent's own
// enters trigger finds them already there. Until now only a spent
// NextSpellPromise wrote it, as the spell became cast. This file is the
// second writer: an effect resolving while the spell waits on the stack
// (a cast trigger, CR 603.2, waits above the spell it watched) gives
// that spell an "enters with" clause, which is a replacement effect on
// the entry (CR 614.1c).
//
// THE MARK LIVES ON THE OBJECT. So it ends with it (CR 400.7): a spell
// countered, or one that left the stack and came back, is a new object
// and has no mark, and the counters are not given to anything else. The
// writer is told the object it means (an ObjectRef, its CR 400.7 epoch
// stamped when the trigger fired) and refuses a different one.
//
// NOT A COPY. A copy of a permanent spell becomes a token (CR 608.3f),
// and a mark on the original is an effect on that object, not a
// copiable value (CR 707.2), so the copy is built without it
// (spell_copy.go) and the writer refuses to mark one directly: an
// effect that names "that creature" names the spell that was cast, and
// a copy was not cast (CR 707.10).
//
// PLAIN DATA. The mark is the same PromisedCounter struct the promise
// writes, already in the snapshot and the clone; nothing here adds a
// field or a func route from Game.

// MarkSpellEntersWithCountersForEffect gives the spell `spell` on the
// stack "enters with `mark.Count` additional `mark.Kind` counters on
// it". It reports false, and writes nothing, when the ref is not a
// spell still on the stack as that same object (it resolved, was
// countered, or left and came back), when it is a copy, or when the
// mark puts no counters on.
//
// The epoch is checked only when the ref carries one the caller read
// (a non-negative Epoch): a writer running AS the spell becomes cast
// (#4679's cast permission) holds the live item and may pass the ID
// with the card's current epoch.
//
// The *ForEffect surface: caller must hold g.mu (write).
func (g *Game) MarkSpellEntersWithCountersForEffect(spell ObjectRef, mark PromisedCounter) bool {
	if mark.Count <= 0 || mark.Kind == "" {
		return false
	}
	item := g.StackMeta[spell.ID]
	card := g.spellCardOnStackLocked(spell.ID)
	if item == nil || card == nil || item.Kind != StackItemSpell || item.IsCopy {
		return false
	}
	if spell.Epoch >= 0 && card.ObjectEpoch != spell.Epoch {
		return false
	}
	addPromisedCounterLocked(item, mark)
	return true
}

// addPromisedCounterLocked appends one mark to the item, onto a fresh
// backing array: an undo snapshot shares the old one
// (sweepPlayerStaticsLocked's reason). The one writer of
// StackItem.PromisedCounters, shared by the promise spend and the
// effect API above.
//
// Caller must hold g.mu (write).
func addPromisedCounterLocked(item *StackItem, mark PromisedCounter) {
	item.PromisedCounters = append(copyPromisedCounters(item.PromisedCounters), mark)
}
