package testsav

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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

// --- the seam -----------------------------------------------------

// The event is the crew ability RESOLVING (CR 702.122e): nothing while
// the ability waits on the stack, one event however many creatures paid,
// and Amount counts the object's resolutions this turn.
func TestBecameCrewedFiresOncePerCrewAbilityThatResolves(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	copter := s.PushVehicle("Smuggler's Copter", effectstest.SmugglersCopterOracle, 3, 3)
	a := s.PushCreature("A", "Creature — Human", 1, 1)
	b := s.PushCreature("B", "Creature — Human", 1, 1)
	c := s.PushCreature("C", "Creature — Human", 1, 1)

	s.CrewOnly(t, copter, a, b)
	if n := len(effectstest.CrewEventsThisTurn(s.G, game.EventBecameCrewed)); n != 0 {
		t.Fatalf("%d became-crewed events while the crew ability is still on the stack", n)
	}
	if n := len(effectstest.CrewEventsThisTurn(s.G, game.EventCrewedVehicle)); n != 2 {
		t.Fatalf("%d crewed-vehicle events, want one per creature tapped (2)", n)
	}
	effectstest.PassPriorityAroundTable(t, s.G)
	evs := effectstest.CrewEventsThisTurn(s.G, game.EventBecameCrewed)
	if len(evs) != 1 {
		t.Fatalf("%d became-crewed events for one crew ability paid by two creatures, want 1", len(evs))
	}
	if ev := evs[0]; ev.CardID != copter || ev.Source != copter || ev.Target != copter || ev.Actor != s.Me.ID || ev.Amount != 1 {
		t.Errorf("event = %+v, want the Vehicle, its controller and Amount 1", ev)
	}

	s.CrewAndResolve(t, copter, 0, c)
	evs = effectstest.CrewEventsThisTurn(s.G, game.EventBecameCrewed)
	if len(evs) != 2 || evs[1].Amount != 2 {
		t.Fatalf("after a second crew: %+v, want a second event with Amount 2", evs)
	}
	if n := becameCrewedCount(s.G, effectstest.ObjectOf(t, s.G, copter)); n != 2 {
		t.Errorf("the tally counts %d, want 2", n)
	}
}

// The count is carried across a restore, and so is a crew ability on the
// stack: it still says "became crewed" when it resolves in the restored
// game.
func TestBecameCrewedSurvivesARestore(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	copter := s.PushVehicle("Smuggler's Copter", effectstest.SmugglersCopterOracle, 3, 3)
	a := s.PushCreature("A", "Creature — Human", 1, 1)
	b := s.PushCreature("B", "Creature — Human", 1, 1)
	s.CrewAndResolve(t, copter, 0, a)
	s.CrewOnly(t, copter, b)
	ref := effectstest.ObjectOf(t, s.G, copter)

	restored := effectstest.RestoreRoundTrip(t, s.G, true)
	if n := becameCrewedCount(restored, ref); n != 1 {
		t.Fatalf("restored tally counts %d, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, restored)
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
	s := effectstest.NewCrewScene(t)
	copter := s.PushVehicle("Smuggler's Copter", effectstest.SmugglersCopterOracle, 3, 3)
	a := s.PushCreature("A", "Creature — Human", 1, 1)
	s.CrewAndResolve(t, copter, 0, a)
	ref := effectstest.ObjectOf(t, s.G, copter)
	effectstest.AdvancePastCleanupOf(t, s.G, s.G.Turn.ActiveSeat)
	if n := becameCrewedCount(s.G, ref); n != 0 {
		t.Errorf("the tally still counts %d after the turn ended", n)
	}
}
