package game

import (
	"testing"

	"github.com/google/uuid"
)

// pay_unless_cause_test.go — #4279. A pay-unless does not stop the
// table, so a keyed one can be answered after the resolution that asked
// has closed and Game.resolving has moved on. Its decline must stamp
// the cause the prompt carries, not whatever the slot holds then.

// payUnlessDiscardCause is a test card's "target player discards a card
// unless they pay {2}": the decline discards the card the prompt carries.
// It reads only what the prompt hands it.
var payUnlessDiscardCause = RegisterPayUnlessElse("test/payunless-discard-cause", func(g *Game, r PayUnlessAnswer) error {
	return g.discardCardsLocked(r.Chooser, r.Carry, discardOptions{cause: DiscardCauseEffect})
})

// openPayUnlessDiscard has seat 1's spell resolve and ask seat 0 to pay
// {2} or discard, then closes the resolution and walks one step on, so
// the slot is gone. Returns the table, the stack item and the card
// seat 0 would discard.
func openPayUnlessDiscard(t *testing.T) (g *Game, item *StackItem, hand uuid.UUID, promptID uuid.UUID) {
	t.Helper()
	g = newActiveGame(t)
	payer, caster := g.Seats[0], g.Seats[1]
	h := NewCard("Spare", payer.ID)
	payer.Hand.PushTop(h)
	item = &StackItem{ID: uuid.New(), Controller: caster.ID, SourceCardID: uuid.New()}
	g.WithWriteLock(func() {
		g.beginResolvingLocked(item)
		g.beginResolutionLocked()
		if err := g.QueuePayUnlessKeyedForEffect(payer.ID, item.SourceCardID, "{2}", "Pay {2}?", payUnlessDiscardCause, []uuid.UUID{h.InstanceID}); err != nil {
			t.Fatalf("QueuePayUnlessKeyedForEffect: %v", err)
		}
		g.endResolutionLocked()
		g.holdForOpenResolutionLocked()
	})
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoicePayUnless {
			promptID = c.ID
		}
	}
	if promptID == uuid.Nil {
		t.Fatal("no pay-unless prompt queued")
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	g.WithWriteLock(func() {
		if g.resolutionOpen || g.resolving != nil {
			t.Fatalf("the resolution is still open after the step advance (open=%v)", g.resolutionOpen)
		}
	})
	return g, item, h.InstanceID, promptID
}

func discardEventOf(g *Game, player uuid.UUID) *Event {
	for i := range g.Events {
		if g.Events[i].Kind == EventDiscardCard && g.Events[i].Actor == player {
			return &g.Events[i]
		}
	}
	return nil
}

func TestKeyedPayUnlessDeclineKeepsItsCauseAfterTheStepAdvanced(t *testing.T) {
	g, item, hand, id := openPayUnlessDiscard(t)
	payer := g.Seats[0]
	// A different item is parked in the slot by the time the answer
	// comes in: it must not become the cause.
	other := &StackItem{ID: uuid.New(), Controller: payer.ID}
	g.WithWriteLock(func() { g.resolving = &resolvingItem{item: other} })

	if err := g.ResolvePayUnless(id, payer.ID, false); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	if payer.Hand.Contains(hand) {
		t.Fatal("the decline did not discard")
	}
	ev := discardEventOf(g, payer.ID)
	if ev == nil {
		t.Fatal("no discard event")
	}
	if ev.Cause != MoveCauseEffect || ev.CauseController != item.Controller || ev.CauseItem != item.ID {
		t.Errorf("discard cause = (%q, %s, %s), want the asking spell (effect, %s, %s)",
			ev.Cause, ev.CauseController, ev.CauseItem, item.Controller, item.ID)
	}
	if !DiscardCausedByOpponentOf(*ev, payer.ID) {
		t.Error("the discard is not read as an opponent's spell causing it")
	}
	g.WithWriteLock(func() {
		if g.carriedCause != nil {
			t.Error("the carried cause outlived the branch")
		}
	})
}

// "Spells and abilities your opponents control can't cause you to
// discard" (Tamiyo) still stops the decline, with the slot long gone.
func TestKeyedPayUnlessDeclineIsStillGatedByACantCauseDiscard(t *testing.T) {
	withProtection(t)
	g, _, hand, id := openPayUnlessDiscard(t)
	payer := g.Seats[0]
	shield := NewCard("Shield", payer.ID)
	shield.TypeLine = "Enchantment"
	shield.OracleID = protectedKey
	g.Battlefield.PushTop(shield)

	if err := g.ResolvePayUnless(id, payer.ID, false); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	if !payer.Hand.Contains(hand) {
		t.Fatal("an opponent's pay-unless decline discarded a protected player's card")
	}
}

// The same prompt, saved and restored with the resolution long closed,
// still stamps the cause.
func TestKeyedPayUnlessDeclineCauseSurvivesARestore(t *testing.T) {
	g, item, hand, id := openPayUnlessDiscard(t)
	_, g = roundTrip(t, g)
	payer := g.Seats[0]

	if err := g.ResolvePayUnless(id, payer.ID, false); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	if payer.Hand.Contains(hand) {
		t.Fatal("the decline did not discard after the restore")
	}
	ev := discardEventOf(g, payer.ID)
	if ev == nil || ev.CauseController != item.Controller || ev.CauseItem != item.ID {
		t.Fatalf("restored decline's discard event = %+v, want the asking spell", ev)
	}
}

// With no resolution open and no cause handed in, nothing is resolving.
func TestResolutionCauseIsEmptyOnceTheResolutionClosed(t *testing.T) {
	g := newActiveGame(t)
	g.WithWriteLock(func() {
		g.resolving = &resolvingItem{item: &StackItem{ID: uuid.New(), Controller: g.Seats[1].ID}}
		g.resolutionOpen = false
		if c := g.resolutionCauseLocked(); c.Kind != "" {
			t.Errorf("a closed resolution gave cause %+v", c)
		}
		g.resolutionOpen = true
		if c := g.resolutionCauseLocked(); c.Kind != MoveCauseEffect {
			t.Errorf("an open resolution gave cause %+v", c)
		}
		err := g.withCarriedCauseLocked(carriedCause{MoveCause: MoveCause{Kind: MoveCauseEffect, Controller: g.Seats[0].ID}}, func() error {
			g.resolutionOpen = false
			if c := g.resolutionCauseLocked(); c.Controller != g.Seats[0].ID {
				t.Errorf("a carried cause was ignored: %+v", c)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	})
}
