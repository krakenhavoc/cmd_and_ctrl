package game

// entry_counters.go — "this permanent enters with N counters on it",
// where N is read from the spell that became it (CR 614.1c, #1002).
//
// WHY IT IS A DECLARATION AND NOT AN OnResolve. "This creature enters
// with X +1/+1 counters on it" is a replacement effect (CR 614.1c):
// the counters are part of the ENTRY, so every other replacement in
// the CR 616 window gets to see and modify them, the permanent's own
// ETB trigger finds them already there, and "whenever one or more
// counters are put on a permanent you control" fires, because by then
// there IS a permanent.
//
// Thirteen catalogued cards used to put them on in OnResolve instead —
// a beat before the card left the stack — and each declared the same
// player-facing caveat with the same stated reason: that an entry
// replacement cannot see the X announced for the spell, because the
// stack item is gone by the time the entry pipeline runs. The reason
// was not true. resolveTopOfStackLocked builds the entry
// ReplacementEvent with `stackItem: item` and hands that same item to
// applyAltCostEntryCountersLocked on the next line, which is how
// escape's "this creature escapes with a +1/+1 counter on it" has
// ridden the pipeline since S29. X is readable at exactly that line,
// and this file is its sibling for the card's own printed clause.
//
// THE SHAPE, in three decisions.
//
//  1. THE CATALOG DECLARES, THE ENGINE READS. A card says what it
//     says — XCounters("+1/+1"), SunburstCounters("+1/+1") — and the
//     count is a pure function of CastCounts. The catalog never
//     touches a StackItem for this, which is what the old comment in
//     cards/effects/mana_spent.go said would be needed and is the
//     reason sunburst went the OnResolve road too.
//
//  2. ONE VOCABULARY, NAMED. CastCounts is the whole list of things
//     a CR 614.1c clause is allowed to read off the announcement, and
//     it is short on purpose: the announced X, the number of times the
//     spell was kicked, and the colours of mana spent on it. Anything
//     else a future card needs is a documented field here rather than
//     a card reaching into PaidCost on its own.
//
//  3. ONE SEEDING SITE. applyCastEntryCountersLocked runs once, from
//     the single place a spell's entry event is built, BEFORE the
//     CR 614 pipeline. Not from every entry point: a Hangarback Walker
//     reanimated out of a graveyard was never cast, has no X, and
//     enters with no counters, which is what CR 107.3b says about it.

// CastCounts is what a CR 614.1c "enters with N counters" clause may
// read off the spell that is becoming this permanent.
//
// A value, computed once per entry, rather than the StackItem itself:
// the catalog declares arithmetic over these three numbers and cannot
// reach anything else, so the engine keeps one reader of the
// announcement record instead of thirteen.
type CastCounts struct {
	// X is the value announced for X when the spell was cast
	// (CR 601.2b) — StackItem.XValue. Zero for a permanent that came
	// from a spell with no {X} in its cost, and a real zero for one
	// announced at X=0 (CR 107.3).
	X int

	// Kicked is CR 702.33d's "the number of times it was kicked",
	// counted off PaidCost.OptionalCosts: 1 for an ordinary kicked
	// spell, N for a multikicker paid N times, 0 for an unkicked one.
	// Kicker and multikicker together, because no rules text tells
	// them apart.
	Kicked int

	// ColorsSpent is CR 702.44a's sunburst count: the number of
	// DISTINCT colours of mana spent to cast the spell. Zero for a
	// payment the engine did not take (PaidCost.OnPaper) — an
	// unrecorded payment claims no colours, which is the
	// weaker-than-printed answer ADR 0068 §3 requires of every reader
	// of that record.
	ColorsSpent int
}

// EntryCountersFromCast is one printed "this permanent enters with N
// counters on it" clause whose N is read from the cast (CR 614.1c).
//
// Declared on effects.Spec.EntersWithCountersFromCast, projected onto
// CardDef at Register, and seeded onto the entry event by
// applyCastEntryCountersLocked. A count of zero or less puts no
// counters on, so a 0/0 cast for X=0 enters as the printed 0/0 it is
// and meets CR 704.5f at the next state-based check (#691).
type EntryCountersFromCast struct {
	// Kind is the counter's name — "+1/+1", "charge", "fire". The
	// same name AddCounterForEffect and ReplacementEvent.
	// EntersWithCounters use.
	Kind string

	// Count is N. Pure arithmetic over the announcement: it is called
	// once, during the entry, with no access to the game state,
	// because CR 614.1c is a fact about the spell and not about the
	// board.
	Count func(cast CastCounts) int
}

// CatalogEntersWithCountersFromCast is the catalog hook the effects
// package wires at init, mirroring CatalogReplacements and its
// neighbours. Nil (no catalog) means no card declares one.
var CatalogEntersWithCountersFromCast func(oracleID string) []EntryCountersFromCast

// EntersWithCountersFromCastFor returns the CR 614.1c clauses the
// card `oracleID` declares, or nil.
func EntersWithCountersFromCastFor(oracleID string) []EntryCountersFromCast {
	if CatalogEntersWithCountersFromCast == nil || oracleID == "" {
		return nil
	}
	return CatalogEntersWithCountersFromCast(oracleID)
}

// castCountsFor reads the three numbers a CR 614.1c clause may use off
// one resolving announcement. The ONE place the entry pipeline reads
// the paid-cost record.
//
// `card` is needed alongside the item because "the number of times it
// was kicked" is positions in the CARD's OptionalCosts slice, and only
// the card knows which position is kicker (ADR 0073 §5).
func castCountsFor(card Card, item *StackItem) CastCounts {
	if item == nil {
		return CastCounts{}
	}
	return CastCounts{
		X:           item.XValue,
		Kicked:      KickedTimesPaid(card, item.Paid.OptionalCosts),
		ColorsSpent: item.Paid.ColorsSpentCount(),
	}
}

// applyCastEntryCountersLocked folds a permanent spell's printed
// "this permanent enters with N counters on it" (CR 614.1c) into its
// ENTRY event, before the CR 614 pipeline runs.
//
// The sibling of applyAltCostEntryCountersLocked, and deliberately the
// line next to it: one of them is the clause an ALTERNATIVE COST
// attaches to the entry ("this creature escapes with a +1/+1 counter
// on it") and the other is the clause the CARD prints. Both are read
// off the resolving StackItem, both are seeded before
// applyReplacementsLocked, and both therefore compose — a card that
// printed both would get both counts in one entry event.
//
// Seeding BEFORE the pipeline is the whole point:
//
//   - Another entry replacement in the same CR 616 window can read and
//     rewrite ev.EntersWithCounters.
//   - The counters land on the PERMANENT, after the move and before
//     EventETB, so the card's own enters trigger sees the finished
//     creature and a "whenever one or more counters are put on a
//     permanent you control" payoff finally fires — which it could not
//     while the counters went onto a card still on the stack.
//   - Doubling Season and Hardened Scales still apply, because the
//     settled map is drained through AddCounterForEffect and opens the
//     ordinary RepEventCounter window for each kind.
//
// A nil item is every entry that is not a resolving spell — a land
// play, a search, a reanimation, a blink, a token — and seeds nothing.
// That is CR 107.3b: a Hangarback Walker put onto the battlefield had
// no X announced for it, so X is 0 and it enters with no counters.
//
// Caller must hold g.mu.
func (g *Game) applyCastEntryCountersLocked(ev *ReplacementEvent, card Card, item *StackItem) {
	if ev == nil || item == nil {
		return
	}
	clauses := EntersWithCountersFromCastFor(CatalogKey(card))
	if len(clauses) == 0 {
		return
	}
	cast := castCountsFor(card, item)
	for _, clause := range clauses {
		if clause.Count == nil {
			continue
		}
		if n := clause.Count(cast); n > 0 {
			ev.AddCounterAtETB(clause.Kind, n)
		}
	}
}
