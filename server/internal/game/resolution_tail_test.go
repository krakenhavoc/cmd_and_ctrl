package game

import (
	"testing"

	"github.com/google/uuid"
)

// resolution_tail_test.go — #3613, CR 608.2c / 608.2n: a resolution
// that asks something finishes after the answer (resolution_tail.go).
// The card-level tests (Council's Judgment, fabricate) live in the
// effects package; these pin the engine shape with a probe spell.

// pauseOnConfirm is a probe effect that asks one yes/no and runs then
// on "yes".
func pauseOnConfirm(chooser uuid.UUID, then func(*Game) error) func(*Game, *StackItem) error {
	return func(g *Game, item *StackItem) error {
		queueConfirmHooked(g, ConfirmPrompt{
			Chooser:  chooser,
			Source:   item.SourceCardID,
			Question: "test — pause here",
		}, then, nil)
		return nil
	}
}

// eventsAbout counts the events of kind naming card id, and reports
// whether any EventResolve for it is LogDeferred.
func eventsAbout(g *Game, kind EventKind, id uuid.UUID) (n int, deferred bool) {
	for _, e := range g.Events {
		if e.Kind != kind || e.CardID != id {
			continue
		}
		n++
		deferred = deferred || e.LogDeferred
	}
	return n, deferred
}

func answerProbe(t *testing.T, g *Game, chooser uuid.UUID) {
	t.Helper()
	c := latestChoiceOfKindForTest(g, PendingChoiceConfirm)
	if c == nil {
		t.Fatalf("the probe's prompt is not open: %+v", g.PendingChoices)
	}
	if err := g.ResolveConfirm(c.ID, chooser, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
}

func TestAPausedSpellWaitsOnTheStackAndFinishesOnceAcrossAnUndo(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushResolvingSpell(t, g, me.ID, pauseOnConfirm(me.ID, nil))
	resolveUntilPrompt(t, g)

	if !g.Stack.Contains(id) || me.Graveyard.Contains(id) {
		t.Fatal("CR 608.2n: the spell waits on the stack for its own prompt")
	}
	if g.HeldResolutionItemForView() == nil {
		t.Error("the view is not told which item is waiting")
	}
	if n, deferred := eventsAbout(g, EventResolve, id); n != 1 || !deferred {
		t.Errorf("EventResolve: %d, deferred %v; want one, with its log line deferred", n, deferred)
	}
	if n, _ := eventsAbout(g, EventResolutionFinished, id); n != 0 {
		t.Error("the resolution finished before the answer")
	}

	snap := g.Clone()
	answerProbe(t, g, me.ID)
	g.RestoreFrom(snap)
	if !g.Stack.Contains(id) || g.HeldResolutionItemForView() == nil {
		t.Fatal("the undo did not restore the spell waiting on its prompt")
	}

	answerProbe(t, g, me.ID)
	if g.Stack.Contains(id) || !me.Graveyard.Contains(id) {
		t.Fatal("the answer finishes the resolution: the spell goes to its owner's graveyard")
	}
	if g.HeldResolutionItemForView() != nil {
		t.Error("the finished item is still held")
	}
	if n, _ := eventsAbout(g, EventResolutionFinished, id); n != 1 {
		t.Errorf("EventResolutionFinished: %d, want exactly one", n)
	}
}

// A spell whose answer moves the spell itself (#489) is not routed a
// second time by the tail.
func TestAPausedSpellThatMovesItselfIsNotRoutedAgain(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	var id uuid.UUID
	id = pushResolvingSpell(t, g, me.ID, pauseOnConfirm(me.ID, func(g *Game) error {
		return g.ExileCardForEffect(id)
	}))
	resolveUntilPrompt(t, g)
	answerProbe(t, g, me.ID)
	if !g.Exile.Contains(id) || me.Graveyard.Contains(id) || g.Stack.Contains(id) {
		t.Errorf("the spell exiled itself and must stay in exile (exile %v, graveyard %v)",
			g.Exile.Contains(id), me.Graveyard.Contains(id))
	}
}

// A resolution that asks nothing is unchanged: the "resolved" line
// stays where EventResolve is, and nothing is parked.
func TestASpellThatAsksNothingKeepsItsResolveLine(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushResolvingSpell(t, g, me.ID, func(*Game, *StackItem) error { return nil })
	passUntilPrompt(t, g)
	if !me.Graveyard.Contains(id) {
		t.Fatal("the spell resolved to the graveyard")
	}
	if n, deferred := eventsAbout(g, EventResolve, id); n != 1 || deferred {
		t.Errorf("EventResolve: %d, deferred %v; want one, not deferred", n, deferred)
	}
	if n, _ := eventsAbout(g, EventResolutionFinished, id); n != 0 {
		t.Error("a resolution that never paused emitted EventResolutionFinished")
	}
	if g.HeldResolutionItemForView() != nil {
		t.Error("a resolution that never paused left an item held")
	}
}
