package game

import (
	"testing"

	"github.com/google/uuid"
)

// entry_counters_if_type_test.go — #4193: a batch entry's per-card
// "an additional <counter> on it if it's a <type>" clause, and an entry's
// loyalty counters on top of the printed starting loyalty (CR 306.5b).

func putFromGraveyardTogetherForTest(t *testing.T, g *Game, opts ZoneEntryOptions, ids ...uuid.UUID) []uuid.UUID {
	t.Helper()
	entries := make([]BatchEntry, 0, len(ids))
	for _, id := range ids {
		entries = append(entries, BatchEntry{CardID: id, From: ZoneGraveyard})
	}
	var entered []uuid.UUID
	g.mu.Lock()
	err := g.PutOntoBattlefieldTogetherKeyedForEffect(entries, opts, hookTail(func(_ *Game, o TailOutcome) error {
		in := o.IDs
		entered = in
		return nil
	}))
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("PutOntoBattlefieldTogetherKeyedForEffect: %v", err)
	}
	return entered
}

var semestersEndClauses = []EntryCountersIfType{
	{CardType: "creature", Counter: CounterPlusOne, N: 1},
	{CardType: "planeswalker", Counter: CounterLoyalty, N: 1},
}

// Each card of ONE entry gets the clauses its own types match: the
// creature a +1/+1 counter, the planeswalker a loyalty counter on top of
// its printed 3, the creature planeswalker both, the artifact nothing.
func TestBatchEntryCountersIfTypeArePerCard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID}
	walker := Card{InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		StartingLoyalty: 3, Owner: me.ID, Controller: me.ID}
	both := Card{InstanceID: uuid.New(), Name: "Creature Walker", TypeLine: "Legendary Creature Planeswalker — Test",
		Power: 2, Toughness: 2, StartingLoyalty: 4, Owner: me.ID, Controller: me.ID}
	relic := Card{InstanceID: uuid.New(), Name: "Relic", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID}
	for _, c := range []Card{bear, walker, both, relic} {
		me.Graveyard.PushTop(c)
	}
	clauses := append([]EntryCountersIfType(nil), semestersEndClauses...)

	entered := putFromGraveyardTogetherForTest(t, g, ZoneEntryOptions{EntersWithCountersIfType: clauses},
		bear.InstanceID, walker.InstanceID, both.InstanceID, relic.InstanceID)
	if len(entered) != 4 {
		t.Fatalf("entered %v, want all four", entered)
	}
	for _, want := range []struct {
		id      uuid.UUID
		plusOne int
		loyalty int
	}{
		{bear.InstanceID, 1, 0},
		{walker.InstanceID, 0, 4},
		{both.InstanceID, 1, 5},
		{relic.InstanceID, 0, 0},
	} {
		c := battlefieldCardForTest(t, g, want.id)
		if c.Counters[CounterPlusOne] != want.plusOne || c.Counters[CounterLoyalty] != want.loyalty {
			t.Errorf("%s counters %v, want +1/+1 %d and loyalty %d", c.Name, c.Counters, want.plusOne, want.loyalty)
		}
	}
	if len(clauses) != 2 || clauses[0] != semestersEndClauses[0] || clauses[1] != semestersEndClauses[1] {
		t.Errorf("the options' clauses were written: %v", clauses)
	}
}

// The clause is judged as the card would be on the battlefield
// (CR 614.12), after the window's other replacements: an existing
// "enters as a copy" answer makes a creature card enter as a
// planeswalker, and it gets the loyalty counter, not the +1/+1 counter.
func TestBatchEntryCountersIfTypeReadTheCopyItEntersAs(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	shapeshifter := Card{InstanceID: uuid.New(), Name: "Shapeshifter", TypeLine: "Creature — Shapeshifter",
		Power: 0, Toughness: 0, Owner: me.ID, Controller: me.ID}
	me.Graveyard.PushTop(shapeshifter)
	walker := Card{Name: "Walker", TypeLine: "Legendary Planeswalker — Test", StartingLoyalty: 3}
	copied := CopiableValuesOf(walker)

	g.mu.Lock()
	g.RegisterReplacementForTest(ReplacementEffect{
		Watches: []EventKind{EventZoneMove},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventMove && ev.CardID == shapeshifter.InstanceID && ev.EntersAsCopyOf == nil
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			v := copied
			ev.EntersAsCopyOf = &v
			return nil
		},
		Label: "test: enters as a copy of a planeswalker",
	})
	g.mu.Unlock()

	putFromGraveyardTogetherForTest(t, g, ZoneEntryOptions{EntersWithCountersIfType: semestersEndClauses},
		shapeshifter.InstanceID)
	c := battlefieldCardForTest(t, g, shapeshifter.InstanceID)
	if !c.IsPlaneswalker() || c.IsCreature() {
		t.Fatalf("entered as %q, want the planeswalker copy", c.TypeLine)
	}
	if c.Counters[CounterPlusOne] != 0 || c.Counters[CounterLoyalty] != 4 {
		t.Errorf("counters %v, want no +1/+1 and loyalty 4 (copied 3 + 1)", c.Counters)
	}
}

// CR 306.5b: a planeswalker given loyalty counters as it enters keeps
// its printed loyalty too — the stamp no longer reads the entry's
// counters as the starting loyalty — and the later stamp in the ETB hook
// does not add it a second time.
func TestEntryLoyaltyCountersAddToThePrintedLoyalty(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	walker := Card{InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		StartingLoyalty: 3, Owner: me.ID, Controller: me.ID}
	me.Graveyard.PushTop(walker)
	putFromGraveyardTogetherForTest(t, g, ZoneEntryOptions{EntersWithCounters: map[string]int{CounterLoyalty: 2}},
		walker.InstanceID)
	if got := battlefieldCardForTest(t, g, walker.InstanceID).Counters[CounterLoyalty]; got != 5 {
		t.Errorf("loyalty %d, want 5 (printed 3 + 2)", got)
	}

	// A walker with no printed loyalty enters with only what it was given.
	zero := Card{InstanceID: uuid.New(), Name: "Zero", TypeLine: "Legendary Planeswalker — Test",
		Owner: me.ID, Controller: me.ID}
	me.Graveyard.PushTop(zero)
	putFromGraveyardTogetherForTest(t, g, ZoneEntryOptions{EntersWithCounters: map[string]int{CounterLoyalty: 2}},
		zero.InstanceID)
	if got := battlefieldCardForTest(t, g, zero.InstanceID).Counters[CounterLoyalty]; got != 2 {
		t.Errorf("no printed loyalty: loyalty %d, want 2", got)
	}
}

// A member a replacement sends somewhere else never enters, and is
// given nothing (CR 614.6).
func TestBatchEntryCountersIfTypeSkipARedirectedMember(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID}
	me.Graveyard.PushTop(bear)
	ev := &ReplacementEvent{Kind: RepEventMove, CardID: bear.InstanceID, OldZone: ZoneGraveyard, NewZone: ZoneExile}
	g.mu.Lock()
	g.seedEntryCountersIfTypeLocked(semestersEndClauses, ev)
	g.mu.Unlock()
	if len(ev.EntersWithCounters) != 0 {
		t.Errorf("a redirected entry was given %v", ev.EntersWithCounters)
	}
}
