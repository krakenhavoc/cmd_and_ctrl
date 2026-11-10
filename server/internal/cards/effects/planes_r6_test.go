package effects

import (
	"testing"
)

// planes_r6_test.go — Planechase plane cards, redo 6 (ADR 0216 slice 4,
// #4870), on the real Planechase table of planes_harness_test.go.

var (
	planeIzzetSteamMaze = planeFixture{"e5f8b2bb-e9bb-4a0a-aa88-7fa0299f52a9", "Izzet Steam Maze", "Plane — Ravnica"}
)

func TestIzzetSteamMazeCopiesAnyPlayersInstantOrSorcery(t *testing.T) {
	requirePlaneRegisteredFull(t, planeIzzetSteamMaze)
	g, plane := newPlaneGame(t, planeIzzetSteamMaze)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Test Ritual", "Sorcery", "", nil)
	if got := planeTriggerControllers(g, plane.InstanceID); len(got) != 1 {
		t.Fatalf("the cast triggered %d plane items, want 1", len(got))
	}
	// Resolve the trigger only: the copy sits above the original.
	for i := 0; i < 8 && itemsFromPlane(g, plane.InstanceID) > 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("pass: %v", err)
		}
	}
	if n := g.Stack.Size(); n != 2 {
		t.Fatalf("stack holds %d spells after the trigger, want the original and its copy", n)
	}
	passPriorityAroundTable(t, g)

	// A creature spell is not copied.
	castCatalogSpell(t, g, "Test Bear", "Creature — Bear", "", nil)
	if n := itemsFromPlane(g, plane.InstanceID); n != 0 {
		t.Errorf("a creature spell triggered the plane (%d items)", n)
	}
	passPriorityAroundTable(t, g)
}

func TestIzzetSteamMazeChaosDiscountsInstantsAndSorceries(t *testing.T) {
	g, plane := newPlaneGame(t, planeIzzetSteamMaze)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	rollBlank(t, g, me.ID)
	if n := itemsFromPlane(g, plane.InstanceID); n != 0 {
		t.Fatalf("a blank triggered the chaos ability")
	}
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if len(nsPromises(me)) != 1 {
		t.Fatalf("promises after chaos: %+v", nsPromises(me))
	}
	ritual := nsHandCard(me, "Costly Ritual", "Sorcery", "{4}{R}")
	if got := nsPrice(t, g, me, ritual).Generic; got != 1 {
		t.Errorf("a {4}{R} sorcery has %d generic after chaos, want 1", got)
	}
	bear := nsHandCard(me, "Costly Bear", "Creature — Bear", "{4}{R}")
	if got := nsPrice(t, g, me, bear).Generic; got != 4 {
		t.Errorf("a creature spell has %d generic, want 4", got)
	}
	// It outlives the plane: planeswalking away keeps the promise.
	g.WithWriteLock(func() { _ = g.PlaneswalkForEffect(me.ID) })
	if len(nsPromises(me)) != 1 {
		t.Error("the promise ended with the plane")
	}
}
