package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// discard_choice_test.go — #651. An effect's discard (Mind Rot,
// looting) is part of the resolving effect (CR 608.2c), so it is a
// PendingChoice the table waits on, not a counter in the cleanup
// step's hand-size map.
//
// The two failures these tests pin, both reproduced on develop before
// the fix:
//
//   - nothing waited for the discard. DiscardPending is not a
//     PendingChoice, so PassPriority and AdvanceStep walked straight
//     past an owed Mind Rot;
//   - entering cleanup ERASED it. populateDiscardPendingLocked resets
//     the map and writes only the active player's hand-size count.

// queueEffectDiscard is the locked-context call a card makes.
func queueEffectDiscard(g *Game, p DiscardPrompt) uuid.UUID {
	var id uuid.UUID
	g.WithWriteLock(func() { id = g.QueueDiscardChoiceForEffect(p) })
	return id
}

// discardPromptFor returns the effect-discard prompt owed by a player.
func discardPromptFor(g *Game, player uuid.UUID) *PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceChooseCards && c.Chooser == player && c.FromPlayer == player {
			return c
		}
	}
	return nil
}

func TestEffectDiscardIsAPromptTheTableWaitsOn(t *testing.T) {
	g := newActiveGame(t)
	victim := g.Seats[1]
	hand := victim.Hand.Size()
	if hand < 3 {
		t.Fatalf("opening hand is %d; this test needs three cards", hand)
	}

	id := queueEffectDiscard(g, DiscardPrompt{Player: victim.ID, N: 2})
	if id == uuid.Nil {
		t.Fatal("a two-card discard against a full hand must queue a prompt")
	}
	c := discardPromptFor(g, victim.ID)
	if c == nil {
		t.Fatal("no discard prompt queued")
	}
	if c.ChooseMin != 2 || c.ChooseMax != 2 {
		t.Errorf("bounds [%d,%d], want [2,2]", c.ChooseMin, c.ChooseMax)
	}
	if len(c.ChooseCards) != hand {
		t.Errorf("the whole hand is on offer: %d candidates, hand %d", len(c.ChooseCards), hand)
	}
	// The cleanup map is CR 514.1's and nothing else touches it now.
	if len(g.DiscardPending) != 0 {
		t.Errorf("an effect discard wrote to the cleanup map: %v", g.DiscardPending)
	}

	// #791's gate, which the discard now rides for free: no verb that
	// moves the table is accepted while the discard is owed.
	if err := g.PassPriority(); !errors.Is(err, ErrChoicePending) {
		t.Errorf("PassPriority: %v, want ErrChoicePending", err)
	}
	if _, err := g.AdvanceStep(); !errors.Is(err, ErrChoicePending) {
		t.Errorf("AdvanceStep: %v, want ErrChoicePending", err)
	}
	if err := g.PassTurn(); !errors.Is(err, ErrChoicePending) {
		t.Errorf("PassTurn: %v, want ErrChoicePending", err)
	}

	picks := []uuid.UUID{victim.Hand.Cards[0].InstanceID, victim.Hand.Cards[1].InstanceID}
	if err := g.ResolveChooseCards(c.ID, victim.ID, picks); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if victim.Hand.Size() != hand-2 {
		t.Errorf("hand %d → %d, want two fewer", hand, victim.Hand.Size())
	}
	for _, pick := range picks {
		if !victim.Graveyard.Contains(pick) {
			t.Errorf("picked card %s is not in the graveyard", pick)
		}
	}
	discards := 0
	for _, ev := range g.Events {
		if ev.Kind == EventDiscardCard && ev.Actor == victim.ID {
			if ev.OldZone != ZoneHand || ev.NewZone != ZoneGraveyard {
				t.Errorf("EventDiscardCard zones %s → %s", ev.OldZone, ev.NewZone)
			}
			discards++
		}
	}
	if discards != 2 {
		t.Errorf("EventDiscardCard x %d, want 2 — the discard triggers watch it", discards)
	}
	if discardPromptFor(g, victim.ID) != nil {
		t.Error("the prompt is dequeued by the answer")
	}
	if err := g.PassPriority(); err != nil {
		t.Errorf("priority passes once the discard is paid: %v", err)
	}
}

// TestCleanupDoesNotEraseAnOwedEffectDiscard is the second half of the
// bug: entering cleanup rebuilds DiscardPending from scratch, which
// used to drop a Mind Rot that was still owed. The two obligations now
// live in different places, so both survive — and they are both still
// owed by the same player.
func TestCleanupDoesNotEraseAnOwedEffectDiscard(t *testing.T) {
	g := newActiveGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	// Nine cards, so the hand-size discard (CR 514.1, cap 7) owes two.
	for active.Hand.Size() < 9 {
		if err := g.DrawCard(active.ID); err != nil {
			t.Fatalf("DrawCard: %v", err)
		}
	}
	queueEffectDiscard(g, DiscardPrompt{Player: active.ID, N: 2})
	owed := discardPromptFor(g, active.ID)
	if owed == nil {
		t.Fatal("no effect discard queued")
	}

	// Enter cleanup the way the step cursor does. (The gate refuses
	// AdvanceStep while the prompt is open — that is the first half of
	// the fix — so the step-entry hook is run directly here to pin the
	// second half: that the hook does not eat the prompt.)
	g.WithWriteLock(func() {
		g.Turn.Step = StepCleanup
		g.populateDiscardPendingLocked()
	})

	if got := g.DiscardPending[active.ID]; got != 2 {
		t.Errorf("the hand-size discard = %d, want 2", got)
	}
	still := discardPromptFor(g, active.ID)
	if still == nil {
		t.Fatal("cleanup erased the owed effect discard — #651")
	}
	if still.ID != owed.ID || still.ChooseMax != 2 {
		t.Errorf("the owed prompt changed: %v", still)
	}
}

// TestEffectDiscardWithNoCardsQueuesNothingAndStillRunsThen — CR
// 701.8a discards as many as you can, and a prompt with no candidates
// and a floor of one is one nobody can answer. What comes after "then"
// is not conditional on there having been cards to pitch.
func TestEffectDiscardWithNoCardsQueuesNothingAndStillRunsThen(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[1]
	g.WithWriteLock(func() { p.Hand.Cards = nil })

	ran := false
	id := queueEffectDiscard(g, DiscardPrompt{
		Player: p.ID,
		N:      2,
		Then:   func(*Game) error { ran = true; return nil },
	})
	if id != uuid.Nil || discardPromptFor(g, p.ID) != nil {
		t.Error("an empty hand must queue no prompt")
	}
	if !ran {
		t.Error(`"discard your hand, THEN draw three" still draws three`)
	}
	if err := g.PassPriority(); err != nil {
		t.Errorf("nothing is owed, so priority passes: %v", err)
	}
}

// TestEffectDiscardCapsAtTheHandSize — "discard two cards" with one
// card in hand discards that one card, and asks for exactly one.
func TestEffectDiscardCapsAtTheHandSize(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[1]
	g.WithWriteLock(func() { p.Hand.Cards = p.Hand.Cards[:1] })

	queueEffectDiscard(g, DiscardPrompt{Player: p.ID, N: 3})
	c := discardPromptFor(g, p.ID)
	if c == nil {
		t.Fatal("one card in hand still owes a discard")
	}
	if c.ChooseMin != 1 || c.ChooseMax != 1 {
		t.Errorf("bounds [%d,%d], want [1,1] — CR 701.8a, as many as you can", c.ChooseMin, c.ChooseMax)
	}
}

// TestEffectDiscardUpToLeavesTheFloorAtZero — "discard up to two
// cards" keeps "discard nothing" as an answer.
func TestEffectDiscardUpToLeavesTheFloorAtZero(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[1]
	queueEffectDiscard(g, DiscardPrompt{Player: p.ID, N: 2, UpTo: true})
	c := discardPromptFor(g, p.ID)
	if c == nil {
		t.Fatal("no prompt queued")
	}
	if c.ChooseMin != 0 || c.ChooseMax != 2 {
		t.Errorf("bounds [%d,%d], want [0,2]", c.ChooseMin, c.ChooseMax)
	}
	hand := p.Hand.Size()
	if err := g.ResolveChooseCards(c.ID, p.ID, nil); err != nil {
		t.Fatalf("declining an 'up to' discard: %v", err)
	}
	if p.Hand.Size() != hand {
		t.Errorf("hand %d → %d, want untouched", hand, p.Hand.Size())
	}
}

// TestEffectDiscardValidateRefusesASetTheCardForbids — the set-level
// hook is the same one the choose-cards prompt and the search prompt
// carry ("discard two cards unless you discard a creature card").
func TestEffectDiscardValidateRefusesASetTheCardForbids(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[1]
	want := p.Hand.Cards[0].InstanceID

	queueEffectDiscard(g, DiscardPrompt{
		Player:   p.ID,
		N:        1,
		Validate: func(picked []Card) bool { return picked[0].InstanceID == want },
	})
	c := discardPromptFor(g, p.ID)
	if c == nil {
		t.Fatal("no prompt queued")
	}
	if err := g.ResolveChooseCards(c.ID, p.ID, []uuid.UUID{p.Hand.Cards[1].InstanceID}); !errors.Is(err, ErrChoiceSetRejected) {
		t.Errorf("a refused set: %v, want ErrChoiceSetRejected", err)
	}
	if discardPromptFor(g, p.ID) == nil {
		t.Error("a refused set leaves the prompt open to try again")
	}
	if !g.ChooseCardsPickLegalLocked(c, []uuid.UUID{want}) {
		t.Error("the enumerator asks the same hook, so a bot is offered the legal set")
	}
	if err := g.ResolveChooseCards(c.ID, p.ID, []uuid.UUID{want}); err != nil {
		t.Fatalf("the legal set: %v", err)
	}
}

// TestEffectDiscardThenRunsAfterTheCardsAreGone is the rummage order —
// "discard a card, then draw a card". Nothing after "then" may run
// while the prompt is open, and when it does run the cards are already
// in the graveyard.
func TestEffectDiscardThenRunsAfterTheCardsAreGone(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[1]
	pick := p.Hand.Cards[0].InstanceID

	inGraveyardWhenThenRan := false
	ran := 0
	queueEffectDiscard(g, DiscardPrompt{
		Player: p.ID,
		N:      1,
		Then: func(g *Game) error {
			ran++
			inGraveyardWhenThenRan = p.Graveyard.Contains(pick)
			return nil
		},
	})
	if ran != 0 {
		t.Fatal("the rest of the effect must wait for the discard")
	}
	c := discardPromptFor(g, p.ID)
	if c == nil {
		t.Fatal("no prompt queued")
	}
	if err := g.ResolveChooseCards(c.ID, p.ID, []uuid.UUID{pick}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if ran != 1 {
		t.Errorf("Then ran %d times, want 1", ran)
	}
	if !inGraveyardWhenThenRan {
		t.Error("Then ran before the discarded card reached the graveyard")
	}
}

// TestEffectDiscardForAnEliminatedPlayerStrandsNothing — a prompt
// addressed to a seat that has left can never be answered, so it is
// never queued; and a seat eliminated while one is open has it pruned
// by the existing CR 800.4a path rather than holding the table.
func TestEffectDiscardForAnEliminatedPlayerStrandsNothing(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	gone, later := g.Seats[1], g.Seats[2]

	ran := false
	g.WithWriteLock(func() { g.eliminatePlayerLocked(gone) })
	id := queueEffectDiscard(g, DiscardPrompt{
		Player: gone.ID,
		N:      1,
		Then:   func(*Game) error { ran = true; return nil },
	})
	if id != uuid.Nil || discardPromptFor(g, gone.ID) != nil {
		t.Error("a departed seat is not prompted")
	}
	if ran {
		t.Error(`"each player discards, then you draw one per card discarded" draws nothing for a seat that has left`)
	}

	// Queued first, eliminated after: the prompt goes with them.
	queueEffectDiscard(g, DiscardPrompt{Player: later.ID, N: 1})
	if discardPromptFor(g, later.ID) == nil {
		t.Fatal("no prompt queued for a live seat")
	}
	if err := g.PassPriority(); !errors.Is(err, ErrChoicePending) {
		t.Fatalf("the open prompt gates the table: %v", err)
	}
	g.WithWriteLock(func() { g.eliminatePlayerLocked(later) })
	if discardPromptFor(g, later.ID) != nil {
		t.Error("an eliminated seat's prompt must be pruned, not stranded")
	}
	if err := g.PassPriority(); err != nil {
		t.Errorf("the table moves again: %v", err)
	}
}
