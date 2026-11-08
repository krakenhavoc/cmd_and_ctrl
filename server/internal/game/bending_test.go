package game

import (
	"testing"

	"github.com/google/uuid"
)

// #5276: the four bending events and the per-turn record (bending.go).

func bendEventCount(g *Game, kind EventKind, actor uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == kind && ev.Actor == actor {
			n++
		}
	}
	return n
}

// Each event counts on its actor's turn tally, and "all four" needs every
// kind at least once.
func TestBendEventsCountOnTheTurnTally(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, other := g.Seats[0].ID, g.Seats[1].ID
	g.WithWriteLock(func() {
		g.NoteBendForEffect(EventWaterbend, me, uuid.Nil)
		g.NoteBendForEffect(EventWaterbend, me, uuid.Nil)
		g.NoteBendForEffect(EventEarthbend, me, uuid.Nil)
		g.NoteBendForEffect(EventFirebend, me, uuid.Nil)
		g.NoteBendForEffect(EventAirbend, other, uuid.Nil)
		// Not a bending kind, and no player: nothing.
		g.NoteBendForEffect(EventForage, me, uuid.Nil)
		g.NoteBendForEffect(EventAirbend, uuid.Nil, uuid.Nil)
	})
	g.WithWriteLock(func() {
		if n := g.BendsThisTurn(me, EventWaterbend); n != 2 {
			t.Errorf("waterbends %d, want 2", n)
		}
		if n := g.BendsThisTurn(me, EventAirbend); n != 0 {
			t.Errorf("airbends %d, want 0 (another player airbent)", n)
		}
		if g.AllFourBendsThisTurn(me) {
			t.Error("all four with no airbend")
		}
		g.NoteAirbendForEffect(me, uuid.Nil, nil)
		if g.AllFourBendsThisTurn(me) {
			t.Error("an airbend that exiled nothing counted")
		}
		g.NoteAirbendForEffect(me, uuid.Nil, []uuid.UUID{uuid.New()})
		if !g.AllFourBendsThisTurn(me) {
			t.Error("all four not reported after each kind")
		}
	})
	if n := bendEventCount(g, EventForage, me); n != 0 {
		t.Errorf("a non-bending kind was emitted %d times", n)
	}
	if n := bendEventCount(g, EventAirbend, me); n != 1 {
		t.Errorf("%d airbend events for me, want 1", n)
	}
}

// The record is per turn: the next turn starts at zero.
func TestBendTallyResetsWithTheTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	g.WithWriteLock(func() {
		for _, k := range BendKinds {
			g.NoteBendForEffect(k, me, uuid.Nil)
		}
	})
	passTurn(t, g)
	g.WithWriteLock(func() {
		for _, k := range BendKinds {
			if n := g.BendsThisTurn(me, k); n != 0 {
				t.Errorf("%s: %d after the turn ended, want 0", k, n)
			}
		}
	})
}

// CR 701.66b: an earthbend is announced as its return trigger is created,
// for the earthbending player; an earthbend whose land is gone does
// nothing and announces nothing. Earthbending the same land twice is two
// earthbends.
func TestEarthbendIsAnnouncedAsItsReturnIsCreated(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	land := earthbendTestLand(t, g, me, "Forest")
	earthbend(t, g, me, land, 2)
	if n := bendEventCount(g, EventEarthbend, me.ID); n != 1 {
		t.Fatalf("%d earthbend events, want 1", n)
	}
	earthbend(t, g, me, land, 1)
	if n := bendEventCount(g, EventEarthbend, me.ID); n != 2 {
		t.Fatalf("%d earthbend events after a second earthbend, want 2", n)
	}
	g.WithWriteLock(func() {
		if n := g.BendsThisTurn(me.ID, EventEarthbend); n != 2 {
			t.Errorf("earthbends this turn %d, want 2", n)
		}
		if err := g.EarthbendForEffect(me.ID, uuid.Nil, uuid.New(), 3); err != nil {
			t.Fatalf("earthbend of a land that is gone: %v", err)
		}
	})
	if n := bendEventCount(g, EventEarthbend, me.ID); n != 2 {
		t.Errorf("an earthbend of a land that is gone was announced (%d events)", n)
	}
}

// A restore point taken mid-turn keeps the record.
func TestBendTallySurvivesASnapshot(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	g.WithWriteLock(func() {
		g.NoteBendForEffect(EventWaterbend, me, uuid.Nil)
		g.NoteBendForEffect(EventEarthbend, me, uuid.Nil)
		g.NoteBendForEffect(EventFirebend, me, uuid.Nil)
	})
	_, restored := roundTrip(t, g)
	restored.WithWriteLock(func() {
		for _, k := range []EventKind{EventWaterbend, EventEarthbend, EventFirebend} {
			if n := restored.BendsThisTurn(me, k); n != 1 {
				t.Errorf("restored %s: %d, want 1", k, n)
			}
		}
		if restored.AllFourBendsThisTurn(me) {
			t.Error("restored game reports all four with no airbend")
		}
		restored.NoteBendForEffect(EventAirbend, me, uuid.Nil)
		if !restored.AllFourBendsThisTurn(me) {
			t.Error("restored game misses the fourth after an airbend")
		}
	})
}
