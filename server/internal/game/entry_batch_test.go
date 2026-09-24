package game

import (
	"testing"

	"github.com/google/uuid"
)

// entry_batch_test.go — #1322 and #1324: a simultaneous battlefield
// entry that can stop for any card's question, and that can take its
// cards from more than one zone. The catalog half (real shocklands,
// Sword of Hearth and Home, Loyal Warhound) is in
// cards/effects/entry_batch_cards_test.go.

// ebPayLifeToUntap registers a pay-life entry replacement on one card:
// "as this enters, you may pay `cost` life; if you don't, it enters
// tapped" — the shockland shape, without the catalog.
func ebPayLifeToUntap(g *Game, cardID, payer uuid.UUID, cost int) {
	g.RegisterReplacementForTest(ReplacementEffect{
		Watches:       []EventKind{EventZoneMove},
		EntryLifeCost: cost,
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventMove && ev.CardID == cardID && ev.NewZone == ZoneBattlefield
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.EntersTapped = true
			return nil
		},
		Controller: func(*ReplacementEvent, *Game, *Card) uuid.UUID { return payer },
		Label:      "test: pay life or enter tapped",
	})
}

// ebPayLifePrompt is the open pay-life prompt, or nil.
func ebPayLifePrompt(g *Game) *PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceEntryPayLife {
			return c
		}
	}
	return nil
}

func ebLand(owner uuid.UUID, name string) Card {
	return Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Land",
		Owner: owner, Controller: owner,
	}
}

func ebTapped(g *Game, id uuid.UUID) bool {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c.Tapped
		}
	}
	return false
}

// TestBatchEntryAsksEachCardInTurnAndLandsTogether is #1322's shape.
// Two lands put by one effect, each with its own "pay life or enter
// tapped": the first card's question is asked with NOTHING moved, the
// second card's question is asked with still nothing moved, and only
// the second answer lands them — both at once, and the caller's
// continuation runs once, with both.
func TestBatchEntryAsksEachCardInTurnAndLandsTogether(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	a, b := ebLand(me.ID, "Fountain"), ebLand(me.ID, "Crypt")
	me.Library.PushTop(a)
	me.Library.PushTop(b)
	lifeBefore := me.Life

	thenRan := 0
	var entered []uuid.UUID
	g.mu.Lock()
	ebPayLifeToUntap(g, a.InstanceID, me.ID, 2)
	ebPayLifeToUntap(g, b.InstanceID, me.ID, 2)
	err := g.PutCardsFromLibraryOntoBattlefieldThenForEffect(
		[]uuid.UUID{a.InstanceID, b.InstanceID}, LibraryEntryOptions{Controller: me.ID},
		func(_ *Game, in []uuid.UUID) error { thenRan++; entered = in; return nil })
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("put: %v", err)
	}

	first := ebPayLifePrompt(g)
	if first == nil {
		t.Fatal("the first card's pay-life question was not asked: the batch is resumable now")
	}
	if g.Battlefield.Contains(a.InstanceID) || g.Battlefield.Contains(b.InstanceID) {
		t.Fatal("a card entered while the first question was open")
	}
	if thenRan != 0 {
		t.Fatal("the continuation ran before the entry was complete")
	}
	if err := g.ResolveEntryPayLife(first.ID, me.ID, true); err != nil {
		t.Fatalf("pay: %v", err)
	}

	second := ebPayLifePrompt(g)
	if second == nil {
		t.Fatal("the second card's question was not asked after the first was answered")
	}
	if g.Battlefield.Contains(a.InstanceID) {
		t.Fatal("the first card landed on its own: the batch enters together, after every question")
	}
	if thenRan != 0 {
		t.Fatal("the continuation ran with a question still open")
	}
	if err := g.ResolveEntryPayLife(second.ID, me.ID, false); err != nil {
		t.Fatalf("decline: %v", err)
	}

	if !g.Battlefield.Contains(a.InstanceID) || !g.Battlefield.Contains(b.InstanceID) {
		t.Fatal("both cards are on the battlefield once the last question is answered")
	}
	if ebTapped(g, a.InstanceID) {
		t.Error("the first card was paid for and should be untapped")
	}
	if !ebTapped(g, b.InstanceID) {
		t.Error("the second card was declined and should be tapped")
	}
	if me.Life != lifeBefore-2 {
		t.Errorf("life = %d, want %d", me.Life, lifeBefore-2)
	}
	if thenRan != 1 || len(entered) != 2 {
		t.Fatalf("continuation ran %d times with %v; want once with both", thenRan, entered)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("%d prompts left open", len(g.PendingChoices))
	}
}

// TestBatchEntryCombinedCostsMustBePayable is CR 614.12b: "that player
// may not make choices for those effects that would cause the combined
// costs of those effects to not be payable." At 3 life, two pay-2
// entries cannot both be paid. The first payment is made as it is
// chosen, so the second question is asked against 1 life — which
// cannot pay 2 (CR 119.4), so it is not asked at all and that land
// enters tapped.
func TestBatchEntryCombinedCostsMustBePayable(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	me.Life = 3
	a, b := ebLand(me.ID, "Fountain"), ebLand(me.ID, "Crypt")
	me.Library.PushTop(a)
	me.Library.PushTop(b)

	g.mu.Lock()
	ebPayLifeToUntap(g, a.InstanceID, me.ID, 2)
	ebPayLifeToUntap(g, b.InstanceID, me.ID, 2)
	_, err := g.PutCardsFromLibraryOntoBattlefieldForEffect(
		[]uuid.UUID{a.InstanceID, b.InstanceID}, LibraryEntryOptions{Controller: me.ID})
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("put: %v", err)
	}
	first := ebPayLifePrompt(g)
	if first == nil {
		t.Fatal("at 3 life the first payment is affordable and is asked")
	}
	if err := g.ResolveEntryPayLife(first.ID, me.ID, true); err != nil {
		t.Fatalf("pay: %v", err)
	}
	if c := ebPayLifePrompt(g); c != nil {
		t.Fatal("asked for a second 2 life at 1 life: the combined costs are not payable (CR 614.12b)")
	}
	if !g.Battlefield.Contains(a.InstanceID) || !g.Battlefield.Contains(b.InstanceID) {
		t.Fatal("both cards enter")
	}
	if ebTapped(g, a.InstanceID) || !ebTapped(g, b.InstanceID) {
		t.Error("the paid land is untapped and the unpayable one tapped")
	}
	if me.Life != 1 {
		t.Errorf("life = %d, want 1", me.Life)
	}
}

// TestBatchEntryFromTwoZonesIsOneEntry is #1324: a card from exile and
// a card from a library, "put both cards onto the battlefield under
// your control". Both land before either is announced — every zone-move
// and ETB event of the batch comes after both are on the battlefield —
// the exiled card returns as a new object (CR 400.7), and both enter
// under the named controller.
func TestBatchEntryFromTwoZonesIsOneEntry(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	creature := Card{
		InstanceID: uuid.New(), Name: "Warhound", TypeLine: "Creature — Dog",
		Power: 3, Toughness: 1, Owner: me.ID, Controller: opp.ID,
	}
	g.Exile.PushTop(creature)
	land := ebLand(me.ID, "Plains")
	me.Library.PushTop(land)

	// A replacement that reads the board as each entry is evaluated:
	// neither card is on the battlefield while the other's window runs.
	sawOther := false
	g.mu.Lock()
	g.RegisterReplacementForTest(ReplacementEffect{
		Watches: []EventKind{EventZoneMove},
		AppliesTo: func(ev *ReplacementEvent, g *Game, _ *Card) bool {
			if ev.Kind == RepEventMove && ev.NewZone == ZoneBattlefield && g.Battlefield.Size() > 0 {
				sawOther = true
			}
			return false
		},
		Label: "test: board watcher",
	})
	before := len(g.Events)
	var entered []uuid.UUID
	err := g.PutOntoBattlefieldTogetherThenForEffect([]BatchEntry{
		{CardID: creature.InstanceID, From: ZoneExile},
		{CardID: land.InstanceID, From: ZoneLibrary},
	}, ZoneEntryOptions{Controller: me.ID}, func(_ *Game, in []uuid.UUID) error {
		entered = in
		return nil
	})
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("put: %v", err)
	}
	if sawOther {
		t.Error("a card's window saw the other card on the battlefield: the windows run against the pre-entry board")
	}
	if len(entered) != 2 {
		t.Fatalf("entered = %v, want both", entered)
	}
	if entered[0] == creature.InstanceID {
		t.Error("the card from exile returned with its old ID; it is a new object (CR 400.7)")
	}
	if entered[1] != land.InstanceID {
		t.Error("the land keeps its library ID")
	}
	for _, id := range entered {
		c, ok := g.battlefieldCardLocked(id)
		if !ok {
			t.Fatalf("%s is not on the battlefield", id)
		}
		if c.Controller != me.ID {
			t.Errorf("%s entered under %s, want the named controller", c.Name, c.Controller)
		}
	}
	if g.Exile.Contains(creature.InstanceID) || me.Library.Contains(land.InstanceID) {
		t.Error("a card is still in its source zone")
	}
	// Each newcomer is announced once. That the announcements come
	// after BOTH have landed is what the catalog test pins through a
	// real intervening-if (Loyal Warhound under Sword of Hearth and
	// Home), because a trigger harvested off the first ETB is the
	// observer CR 603.6a is about.
	var moves, etbs int
	for _, ev := range g.Events[before:] {
		switch ev.Kind {
		case EventZoneMove:
			moves++
		case EventETB:
			etbs++
		}
	}
	if moves != 2 || etbs != 2 {
		t.Errorf("zone moves %d, ETBs %d; want 2 and 2", moves, etbs)
	}
}

// TestBatchEntryUndoAcrossAQuestionReplays pins the clone contract:
// the batch's cursor moves as its questions are answered, so a snapshot
// taken with the first question open must keep its OWN batch. Answering
// on the live game and then restoring the snapshot has to ask the first
// question again and then the second, not skip straight to the landing.
func TestBatchEntryUndoAcrossAQuestionReplays(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	a, b := ebLand(me.ID, "Fountain"), ebLand(me.ID, "Crypt")
	me.Library.PushTop(a)
	me.Library.PushTop(b)

	g.mu.Lock()
	ebPayLifeToUntap(g, a.InstanceID, me.ID, 2)
	ebPayLifeToUntap(g, b.InstanceID, me.ID, 2)
	_, err := g.PutCardsFromLibraryOntoBattlefieldForEffect(
		[]uuid.UUID{a.InstanceID, b.InstanceID}, LibraryEntryOptions{Controller: me.ID})
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("put: %v", err)
	}
	snap := g.Clone()

	for i := 0; i < 2; i++ {
		c := ebPayLifePrompt(g)
		if c == nil {
			t.Fatalf("live game: question %d missing", i+1)
		}
		if err := g.ResolveEntryPayLife(c.ID, me.ID, false); err != nil {
			t.Fatalf("live answer %d: %v", i+1, err)
		}
	}
	if !g.Battlefield.Contains(a.InstanceID) || !g.Battlefield.Contains(b.InstanceID) {
		t.Fatal("the live game lands both")
	}

	g.WithWriteLock(func() { g.RestoreFrom(snap) })
	meID := g.Seats[0].ID
	if g.Battlefield.Contains(a.InstanceID) || g.Battlefield.Contains(b.InstanceID) {
		t.Fatal("the rewind puts both cards back in the library")
	}
	for i := 0; i < 2; i++ {
		c := ebPayLifePrompt(g)
		if c == nil {
			t.Fatalf("rewound game: question %d missing — the live answers walked the snapshot's batch", i+1)
		}
		if i == 0 && g.Battlefield.Contains(a.InstanceID) {
			t.Fatal("rewound game landed a card before its second question")
		}
		if err := g.ResolveEntryPayLife(c.ID, meID, true); err != nil {
			t.Fatalf("rewound answer %d: %v", i+1, err)
		}
	}
	if !g.Battlefield.Contains(a.InstanceID) || !g.Battlefield.Contains(b.InstanceID) {
		t.Fatal("the replayed answers land both")
	}
	if ebTapped(g, a.InstanceID) || ebTapped(g, b.InstanceID) {
		t.Error("both were paid for on the replay and should be untapped")
	}
}

// TestBatchEntryPromptPrunedWhenItsCardLeaves: the card whose question
// is open is milled out from under it. The prompt goes (the move it
// asks about cannot happen), and the batch is told that card is not
// part of the entry — the rest still lands, and the continuation still
// runs. A batch waiting on a prompt that no longer exists would never
// finish.
func TestBatchEntryPromptPrunedWhenItsCardLeaves(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	a, b := ebLand(me.ID, "Fountain"), ebLand(me.ID, "Forest")
	me.Library.PushTop(b)
	me.Library.PushTop(a)

	thenRan := 0
	var entered []uuid.UUID
	g.mu.Lock()
	ebPayLifeToUntap(g, a.InstanceID, me.ID, 2)
	err := g.PutCardsFromLibraryOntoBattlefieldThenForEffect(
		[]uuid.UUID{a.InstanceID, b.InstanceID}, LibraryEntryOptions{Controller: me.ID},
		func(_ *Game, in []uuid.UUID) error { thenRan++; entered = in; return nil })
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("put: %v", err)
	}
	if ebPayLifePrompt(g) == nil {
		t.Fatal("no question for the first card")
	}

	// Mill the first card (the top of the library) while its question
	// is open.
	g.mu.Lock()
	if err := g.MillNForEffect(me.ID, 1); err != nil {
		g.mu.Unlock()
		t.Fatalf("mill: %v", err)
	}
	g.mu.Unlock()

	if ebPayLifePrompt(g) != nil {
		t.Fatal("the question about a card that left the library is still open")
	}
	if thenRan != 1 {
		t.Fatalf("continuation ran %d times, want 1", thenRan)
	}
	if len(entered) != 1 || entered[0] != b.InstanceID || !g.Battlefield.Contains(b.InstanceID) {
		t.Errorf("entered = %v, want only the Forest", entered)
	}
	if g.Battlefield.Contains(a.InstanceID) {
		t.Error("the milled card was pulled back out of the graveyard")
	}
}
