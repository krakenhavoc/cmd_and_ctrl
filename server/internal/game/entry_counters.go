package game

import (
	"sort"

	"github.com/google/uuid"
)

// entry_counters.go — "this permanent enters with N counters on it",
// where N is read from the spell that became it (CR 614.1c, #1002),
// and the one place the settled map of those counters is drained onto
// the permanent (#1010).
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
//     says — XCounters("+1/+1"), CountersPerKick("charge", 1) — and the
//     count is a pure function of CastCounts. The catalog never
//     touches a StackItem for this, which is what the old comment in
//     cards/effects/mana_spent.go said would be needed and is the
//     reason sunburst went the OnResolve road too.
//
//  2. ONE VOCABULARY, NAMED. CastCounts is the whole list of things
//     a CR 614.1c clause is allowed to read off the announcement, and
//     it is short on purpose: the announced X, the number of times the
//     spell was kicked, the colours of mana spent on it, the total
//     amount of mana spent on it (#1735), and the cards delve exiled
//     to pay for it (ADR 0100 sub-PR 2). Anything else a future card
//     needs is a documented field here rather than a card reaching
//     into PaidCost on its own.
//
//     The same vocabulary is what a catalog hook that runs INSIDE the
//     window and is handed the event may read —
//     EntryCastCountsForEffect, below. CopySelector.Candidates is the
//     first such hook (Mockingbird's "with mana value less than or
//     equal to the amount of mana spent to cast this creature").
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
// the catalog declares arithmetic over these fields and cannot
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

	// KickersPaid is the mana cost of each kicker that was paid, as the
	// card declares it ("{1}{G}", "{W}") — the record behind CR
	// 702.33f's "if it was kicked with its [cost] kicker" (#2360).
	// Read it through KickedWith. Empty for an unkicked spell and for
	// every entry that was not a resolving spell, a CR 707.10 copy
	// included, which paid nothing.
	KickersPaid []string

	// ColorsSpent is CR 702.44a's sunburst count: the number of
	// DISTINCT colours of mana spent to cast the spell. Zero for a
	// payment the engine did not take (PaidCost.OnPaper) — an
	// unrecorded payment claims no colours, which is the
	// weaker-than-printed answer ADR 0068 §3 requires of every reader
	// of that record.
	ColorsSpent int

	// ManaSpent is "the amount of mana spent to cast" the spell
	// (CR 601.2h, #1735): every mana that left the caster's pool to pay
	// the total cost CR 601.2f settled — generic, coloured, X, kicker
	// and any other additional cost paid in mana, the commander tax
	// and every other increase, after every reduction. Mana from a
	// Treasure or a ritual is mana like any other.
	//
	// What it does NOT count, because none of it is mana: the life a
	// Phyrexian symbol was paid with (CR 107.4f), creatures convoke
	// tapped and cards delve exiled (CR 702.51a, CR 702.66a — both pay
	// for a part of the cost "rather than" mana), and any non-mana
	// additional cost.
	//
	// It is PaidCost.Mana's length, read through the one view
	// (ManaSpent.Total), not a number of its own: the record already
	// held the tokens, and a second count is a second chance to
	// disagree with them. So it inherits the record's two absences.
	// A cast for no mana — "without paying its mana cost", a {0}
	// alternative cost, a CR 707.10 copy — is a KNOWN zero. A payment
	// the engine waived (permissive mode, a strict-mode ForceCast,
	// PaidCost.OnPaper) is also zero, the weaker-than-printed answer
	// ADR 0068 §3 requires: "spent nothing we can see" must never
	// read as "spent enough".
	ManaSpent int

	// Delved is CR 607.2q's "cards exiled with it" for a spell with
	// delve (ADR 0100 sub-PR 2): the cards delve exiled to pay for the
	// spell (PaidCost.Delved) that are still in exile as those objects
	// when it enters, as value copies read where they sit. A card that
	// left exile in the meantime is a new object (CR 400.7) and is not
	// here. Murktide Regent's "a +1/+1 counter on it for each instant
	// and sorcery card exiled with it" is CountersPerDelved over this.
	//
	// Nil for a spell that delved nothing, and for a CR 707.10 copy,
	// which paid nothing (PaidCost.Delved is not copied).
	Delved []Card
}

// KickedWith is CR 702.33f's "if it was kicked with its [cost]
// kicker": whether the kicker whose mana is `cost` was paid for this
// spell. False for a cost the card does not declare.
func (c CastCounts) KickedWith(cost string) bool {
	for _, k := range c.KickersPaid {
		if k == cost {
			return true
		}
	}
	return false
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

// castCountsFor reads the numbers a CR 614.1c clause may use off one
// resolving announcement. With the Delved line in castCountsLocked,
// which needs the exile zone and so the game, it is the ONE place the
// entry pipeline reads the paid-cost record.
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
		KickersPaid: kickersPaidFor(CatalogKey(card), item.Paid.OptionalCosts),
		ColorsSpent: item.Paid.ColorsSpentCount(),
		ManaSpent:   item.Paid.ManaSpentCount(),
	}
}

// castCountsLocked is castCountsFor plus the one field that needs the
// game: the delved cards still in exile as the objects delve put
// there (CR 607.2q). The single builder of a full CastCounts, shared
// by the clause seeder below and EntryCastCountsForEffect, so a
// counter clause and a candidate filter can never read two different
// answers off one entry.
//
// Caller must hold g.mu.
func (g *Game) castCountsLocked(card Card, item *StackItem) CastCounts {
	cast := castCountsFor(card, item)
	if item != nil {
		cast.Delved = g.DelvedCardsForEffect(item.Paid.Delved)
	}
	return cast
}

// EntryCastCountsForEffect is CastCounts for the spell an entry event
// is turning into a permanent — the read for a catalog hook that runs
// INSIDE the CR 614 window and is handed the event rather than
// declared as an EntryCountersFromCast clause (#1735).
//
// CopySelector.Candidates is the first such hook: Mockingbird's "enter
// as a copy of any creature on the battlefield with mana value less
// than or equal to the amount of mana spent to cast this creature"
// filters its candidates on ManaSpent. Both evaluations of the
// candidate list — the prompt, and the re-check when the answer
// arrives — see the same figure, because the event carries the
// resolving item across the pause (ReplacementEvent.stackItem) and the
// payment record on it never changes after the cast.
//
// The zero CastCounts for every entry that is not a resolving spell —
// a reanimation, a flicker, a library search, a token, a land play —
// because nothing was cast and nothing was spent: X is 0 (CR 107.3b),
// and a Mockingbird put onto the battlefield can copy only a creature
// with mana value 0. The same zero for a nil event.
//
// The card is read off the stack, where a resolving permanent spell
// still is while its entry is open (and while it is paused on a
// prompt); only the kicker count needs it.
//
// Caller must hold g.mu (read or write).
func (g *Game) EntryCastCountsForEffect(ev *ReplacementEvent) CastCounts {
	if ev == nil || ev.stackItem == nil {
		return CastCounts{}
	}
	card, _ := g.cardInZoneLocked(g.Stack, ev.CardID)
	return g.castCountsLocked(card, ev.stackItem)
}

// EntrySpentForEffect is the whole spend of the spell an entry event
// is turning into a permanent (ADR 0109 §11 decision 1, #1552): the
// same ManaSpent view a resolving spell reads off its own stack item
// and a permanent reads off its Provenance (CR 400.7d), handed to a
// replacement that runs INSIDE the CR 614 window — including one on
// ANOTHER permanent.
//
// CastCounts is the short, named list a printed clause may read off
// its own announcement. This is the other reader: a replacement on
// Coin of Mastery asking how much of a creature's mana came from an
// artifact source (CountFrom(ManaSourceArtifact)), Kalain asking how
// much came from a Treasure, and "if it wasn't cast or no mana was
// spent to cast it" (None) on Freestrider Commando and Primeval
// Spawn. One view, so every question about spent mana is still
// answered in one file (mana_spent.go).
//
// The zero ManaSpent — "nothing was spent, and that is known" — for
// every entry that is not a resolving spell: a reanimation, a
// flicker, a token, a land play. Nothing was cast, so nothing was
// spent, and None answers true, which is exactly what "if it wasn't
// cast or no mana was spent" needs. A CR 707.10 copy of a permanent
// spell is the same zero (it was not cast). A waived payment (strict
// mana off) is unknown: every count is zero and None is false, the
// weaker answer ADR 0068 §3 requires. The same zero for a nil event.
//
// Caller must hold g.mu (read or write).
func (g *Game) EntrySpentForEffect(ev *ReplacementEvent) ManaSpent {
	if ev == nil || ev.stackItem == nil {
		return ManaSpent{}
	}
	return ev.stackItem.Paid.Spent()
}

// KeywordSunburst is CR 702.44's token in canonicalKeywords.
const KeywordSunburst = "sunburst"

// applySunburstLocked seeds CR 702.44a's counters onto the entry event:
// for each instance of sunburst, one counter per colour of mana spent
// to cast the spell — a +1/+1 counter if the object is entering as a
// creature "ignoring any type-changing effects that would affect it",
// a charge counter otherwise. So the creature test reads the PRINTED
// type line, not the layered one.
//
// The instances are the permanent's as it would exist on the
// battlefield (CR 614.12, ADR 0109 owner decision 1), from the entry
// look-ahead (entry_lookahead.go): its printed and deck-imported
// sunburst, the ones an effect gave the SPELL (Lux Artillery, Solar
// Array — CR 400.7a carries them onto the permanent), each counted
// separately (CR 702.44d), and none at all when it would enter under
// an ability-removing effect such as Dress Down.
//
// CR 702.44b: only for an object entering from the stack as a resolving
// spell, and only if coloured mana was spent. The caller passes the
// resolving item; a payment the engine waived (strict mana off) claims
// no colours and adds nothing (ADR 0068 §3).
func (g *Game) applySunburstLocked(ev *ReplacementEvent, card Card, item *StackItem) {
	instances := g.entryLookAheadLocked(ev).sunburst
	if instances == 0 {
		return
	}
	colors := item.Paid.ColorsSpentCount()
	if colors <= 0 {
		return
	}
	kind := CounterCharge
	if _, types, _ := ParseTypeLine(card.TypeLine); containsKeyword(types, "Creature") {
		kind = CounterPlusOne
	}
	ev.AddCounterAtETB(kind, instances*colors)
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
	// #1547: "if this mana is spent to cast a creature spell, that
	// creature enters with an additional +1/+1 counter on it"
	// (Biophagus) — a spend rider on the mana that paid, seeded in the
	// same place and for the same reason as the card's own clause, so
	// the two compose and Doubling Season sees both.
	g.riderEntryCountersLocked(ev, item)
	// ADR 0109 §11 decision 2 and owner decision 1: sunburst is a
	// keyword, counted on the permanent the entry look-ahead says this
	// would be, so a granted instance counts like a printed one, each
	// separately (CR 702.44d), and an ability-removing effect leaves none.
	g.applySunburstLocked(ev, card, item)
	clauses := EntersWithCountersFromCastFor(CatalogKey(card))
	if len(clauses) == 0 {
		return
	}
	cast := g.castCountsLocked(card, item)
	for _, clause := range clauses {
		if clause.Count == nil {
			continue
		}
		if n := clause.Count(cast); n > 0 {
			ev.AddCounterAtETB(clause.Kind, n)
		}
	}
}

// applyEntryCountersLocked puts the settled "enters with" counters
// onto a permanent that has just arrived. It is the ONE drain of
// ReplacementEvent.EntersWithCounters: every battlefield entry site
// calls it, and nothing else ranges that map.
//
// THE ORDER IS THE POINT (#1010). Each kind is placed through
// AddCounterForEffect, which opens its own RepEventCounter window, so
// with two counter KINDS on one entry the drain order is the order
// those windows open, the order a CR 616 ordering prompt inside them
// is asked in, and the order the EventCounterPlaced events land in the
// log. Go randomises map iteration, so a bare `range` made all three
// differ on every run and on every replay of the same game — the one
// place an entry could come out differently for no reason a player
// could point at, in an engine that is otherwise deterministic from
// its event log (Clone, the snapshot, the bot harness and the replay
// tooling all rest on that).
//
// The order taken is the canonical one: counter name, ascending. NOT
// the order the clauses seeded the map in, and not a question put to
// the affected player. CR 616.1 gives that player the choice when two
// replacement EFFECTS would apply to one event; two kinds on one entry
// are not two effects. They are one settled event with two components
// — the CR 616 window that produced them has already closed — and no
// kind's window can change what another kind's window does, so the
// board is identical whichever goes first and the only thing a choice
// would decide is which of two log lines comes first. Canonical for
// the reason proliferate is (sortedCounterKinds, proliferate.go): an
// event log that reorders between runs makes replay diffs unreadable.
//
// Every key is drained, a zero or negative one included, exactly as
// the four bare ranges this replaces did — AddCounterForEffect no-ops
// on zero, and a replacement that turned an entry's counters negative
// keeps whatever meaning it had. That is why this sorts the keys
// itself rather than calling sortedCounterKinds, which drops the
// non-positive cells.
//
// Errors are dropped, as all four call sites dropped them: a counter
// that will not go onto a permanent which has already arrived is not a
// reason to fail the entry.
//
// Caller must hold g.mu.
func (g *Game) applyEntryCountersLocked(cardID uuid.UUID, counters map[string]int) {
	if cardID == uuid.Nil || len(counters) == 0 {
		return
	}
	kinds := make([]string, 0, len(counters))
	for name := range counters {
		kinds = append(kinds, name)
	}
	sort.Strings(kinds)
	for _, name := range kinds {
		_ = g.AddCounterForEffect(cardID, name, counters[name])
	}
}
