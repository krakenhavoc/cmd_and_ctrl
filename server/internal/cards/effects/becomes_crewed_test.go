package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// becomes_crewed_test.go — "whenever this Vehicle becomes crewed" (CR
// 702.122e, #5671): the event a crew ability emits as it resolves, the
// per-turn count behind "for the first time each turn". The two cards
// that wait on it are tested in their own files.

// becameCrewedCount is how many times the Vehicle object became crewed
// this turn, by the turn tally.
func becameCrewedCount(g *game.Game, ref game.ObjectRef) (n int) {
	g.ReadSnapshot(func() {
		for _, r := range g.TurnTally.BecameCrewed {
			if r == ref {
				n++
			}
		}
	})
	return n
}

// crewOnly activates ability 0 (the crew ability) and leaves it on the
// stack.
func (s crewScene) crewOnly(t *testing.T, vehicle uuid.UUID, crewers ...uuid.UUID) {
	t.Helper()
	toMain(t, s.g)
	if err := s.g.ActivateCatalogAbility(s.me.ID, vehicle, 0, game.ActivateAbilityParams{CrewIDs: crewers}); err != nil {
		t.Fatalf("crew: %v", err)
	}
}

// --- the seam -----------------------------------------------------

// The event is the crew ability RESOLVING (CR 702.122e): nothing while
// the ability waits on the stack, one event however many creatures paid,
// and Amount counts the object's resolutions this turn.
func TestBecameCrewedFiresOncePerCrewAbilityThatResolves(t *testing.T) {
	s := newCrewScene(t)
	copter := s.pushVehicle("Smuggler's Copter", smugglersCopterOracle, 3, 3)
	a := s.pushCreature("A", "Creature — Human", 1, 1)
	b := s.pushCreature("B", "Creature — Human", 1, 1)
	c := s.pushCreature("C", "Creature — Human", 1, 1)

	s.crewOnly(t, copter, a, b)
	if n := len(crewEventsThisTurn(s.g, game.EventBecameCrewed)); n != 0 {
		t.Fatalf("%d became-crewed events while the crew ability is still on the stack", n)
	}
	if n := len(crewEventsThisTurn(s.g, game.EventCrewedVehicle)); n != 2 {
		t.Fatalf("%d crewed-vehicle events, want one per creature tapped (2)", n)
	}
	passPriorityAroundTable(t, s.g)
	evs := crewEventsThisTurn(s.g, game.EventBecameCrewed)
	if len(evs) != 1 {
		t.Fatalf("%d became-crewed events for one crew ability paid by two creatures, want 1", len(evs))
	}
	if ev := evs[0]; ev.CardID != copter || ev.Source != copter || ev.Target != copter || ev.Actor != s.me.ID || ev.Amount != 1 {
		t.Errorf("event = %+v, want the Vehicle, its controller and Amount 1", ev)
	}

	s.crewAndResolve(t, copter, 0, c)
	evs = crewEventsThisTurn(s.g, game.EventBecameCrewed)
	if len(evs) != 2 || evs[1].Amount != 2 {
		t.Fatalf("after a second crew: %+v, want a second event with Amount 2", evs)
	}
	if n := becameCrewedCount(s.g, objectOf(t, s.g, copter)); n != 2 {
		t.Errorf("the tally counts %d, want 2", n)
	}
}

// The count is carried across a restore, and so is a crew ability on the
// stack: it still says "became crewed" when it resolves in the restored
// game.
func TestBecameCrewedSurvivesARestore(t *testing.T) {
	s := newCrewScene(t)
	copter := s.pushVehicle("Smuggler's Copter", smugglersCopterOracle, 3, 3)
	a := s.pushCreature("A", "Creature — Human", 1, 1)
	b := s.pushCreature("B", "Creature — Human", 1, 1)
	s.crewAndResolve(t, copter, 0, a)
	s.crewOnly(t, copter, b)
	ref := objectOf(t, s.g, copter)

	restored := restoreRoundTrip(t, s.g, true)
	if n := becameCrewedCount(restored, ref); n != 1 {
		t.Fatalf("restored tally counts %d, want 1", n)
	}
	passPriorityAroundTable(t, restored)
	if n := becameCrewedCount(restored, ref); n != 2 {
		t.Fatalf("after the restored crew ability resolved the tally counts %d, want 2", n)
	}
	var last game.Event
	restored.WithWriteLock(func() {
		for _, ev := range restored.EventsThisTurn() {
			if ev.Kind == game.EventBecameCrewed {
				last = ev
			}
		}
	})
	if last.CardID != copter || last.Amount != 2 {
		t.Errorf("restored event = %+v, want the Vehicle with Amount 2", last)
	}
}

// The count is a fact about the turn.
func TestBecameCrewedCountEndsWithTheTurn(t *testing.T) {
	s := newCrewScene(t)
	copter := s.pushVehicle("Smuggler's Copter", smugglersCopterOracle, 3, 3)
	a := s.pushCreature("A", "Creature — Human", 1, 1)
	s.crewAndResolve(t, copter, 0, a)
	ref := objectOf(t, s.g, copter)
	advancePastCleanupOf(t, s.g, s.g.Turn.ActiveSeat)
	if n := becameCrewedCount(s.g, ref); n != 0 {
		t.Errorf("the tally still counts %d after the turn ended", n)
	}
}
