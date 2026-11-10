package game

import (
	"testing"

	"github.com/google/uuid"
)

// investigate_test.go — #5425, CR 701.16a: investigate is the Clue's
// creation and then EventInvestigate, once per time the action is taken.

func withInvestigateToken(t *testing.T) {
	t.Helper()
	investigateToken.RLock()
	old := investigateToken.build
	investigateToken.RUnlock()
	RegisterInvestigateToken(func() Card {
		c := NewCard("Clue", uuid.Nil)
		c.TypeLine = "Token Artifact — Clue"
		return c
	})
	t.Cleanup(func() { RegisterInvestigateToken(old) })
}

func investigateEvents(g *Game, player uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == EventInvestigate && ev.Actor == player {
			n++
		}
	}
	return n
}

func cluesOf(g *Game, player uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Clue" && c.Controller == player {
			n++
		}
	}
	return n
}

func TestInvestigateCreatesAClueAndEmitsOneEventEachTime(t *testing.T) {
	withInvestigateToken(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	source := uuid.New()
	g.WithWriteLock(func() {
		if err := g.InvestigateForEffect(me.ID, source, 2); err != nil {
			t.Fatalf("InvestigateForEffect: %v", err)
		}
		if err := g.InvestigateForEffect(opp.ID, source, 0); err != nil {
			t.Fatalf("InvestigateForEffect(0): %v", err)
		}
	})
	if got := cluesOf(g, me.ID); got != 2 {
		t.Errorf("%d Clues, want 2", got)
	}
	if got := investigateEvents(g, me.ID); got != 2 {
		t.Errorf("%d investigate events, want 2", got)
	}
	if got := investigateEvents(g, opp.ID); got != 0 {
		t.Errorf("investigating zero times emitted %d events", got)
	}
	// The event follows its Clue: the first event is after the first
	// token and before the second.
	var order []EventKind
	for _, ev := range g.Events {
		if ev.Kind == EventInvestigate || ev.Kind == EventTokenCreated {
			order = append(order, ev.Kind)
		}
	}
	want := []EventKind{EventTokenCreated, EventInvestigate, EventTokenCreated, EventInvestigate}
	if len(order) != len(want) {
		t.Fatalf("events %v, want %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("events %v, want %v", order, want)
		}
	}
	g.WithWriteLock(func() {
		if got := g.InvestigatedThisTurn(me.ID); got != 2 {
			t.Errorf("InvestigatedThisTurn = %d, want 2", got)
		}
		if got := g.InvestigatedThisTurn(opp.ID); got != 0 {
			t.Errorf("the opponent's InvestigatedThisTurn = %d, want 0", got)
		}
	})
}

// A plain Clue creation is not an investigate.
func TestCreatingAClueIsNotInvestigating(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	clue := NewCard("Clue", uuid.Nil)
	clue.TypeLine = "Token Artifact — Clue"
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, clue, 1); err != nil {
			t.Fatalf("CreateTokenForEffect: %v", err)
		}
	})
	if got := investigateEvents(g, me.ID); got != 0 {
		t.Errorf("%d investigate events from a plain Clue, want 0", got)
	}
}

func TestInvestigatedThisTurnResetsWithTheTurn(t *testing.T) {
	withInvestigateToken(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { _ = g.InvestigateForEffect(me.ID, uuid.Nil, 1) })
	seat := g.Turn.ActiveSeat
	for i := 0; i < 60 && g.Turn.ActiveSeat == seat; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.ActiveSeat == seat {
		t.Fatal("the turn never ended")
	}
	g.WithWriteLock(func() {
		if got := g.InvestigatedThisTurn(me.ID); got != 0 {
			t.Errorf("InvestigatedThisTurn = %d on a new turn, want 0", got)
		}
	})
}
