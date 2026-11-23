package game

import (
	"testing"

	"github.com/google/uuid"
)

// exile_and_suspend_test.go — "exile that card with three time counters on
// it instead of putting it into your graveyard as it resolves. Then if the
// exiled card doesn't have suspend, it gains suspend" (Gandalf of the Secret
// Fire, S82 GAP batch 117).

func castSuspendMarked(t *testing.T, g *Game, me *Player) uuid.UUID {
	t.Helper()
	advanceTo(t, g, StepPrecombatMain)
	c := NewCard("Test Sorcery", me.ID)
	c.TypeLine = "Sorcery"
	c.ManaCost = "{1}"
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	var ok bool
	g.WithWriteLock(func() { ok = g.MarkSpellExilesAndSuspendsForEffect(c.InstanceID, 3) })
	if !ok {
		t.Fatal("the spell on the stack could not be marked")
	}
	return c.InstanceID
}

func TestSuspendMarkedSpellIsExiledWithTimeCountersAndGainsSuspend(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := castSuspendMarked(t, g, me)
	resolveTop(t, g)
	if me.Graveyard.Contains(id) {
		t.Error("a marked spell that resolved went to the graveyard")
	}
	c := exiledCardByIDLocked(g, id)
	if c == nil {
		t.Fatal("a marked spell that resolved was not exiled")
	}
	if c.Counters[CounterTime] != 3 {
		t.Errorf("counters = %v, want three time counters", c.Counters)
	}
	if !CardHasSuspend(*c) {
		t.Error("the exiled card did not gain suspend")
	}
}

func TestSuspendMarkedSpellThatIsCounteredGoesToTheGraveyard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := castSuspendMarked(t, g, me)
	if err := g.CounterSpell(id, nil); err != nil {
		t.Fatalf("CounterSpell: %v", err)
	}
	if !me.Graveyard.Contains(id) || g.Exile.Contains(id) {
		t.Error("a countered marked spell must go to the graveyard")
	}
}

func TestSuspendMarkingRefusesAnythingButASpellOnTheStack(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	var ok bool
	g.WithWriteLock(func() { ok = g.MarkSpellExilesAndSuspendsForEffect(uuid.New(), 3) })
	if ok {
		t.Error("an unknown ID was marked")
	}
	id := castSuspendMarked(t, g, me)
	g.WithWriteLock(func() { ok = g.MarkSpellExilesAndSuspendsForEffect(id, 0) })
	if ok {
		t.Error("a zero counter count was accepted")
	}
}

func TestExileSuspendKindRoundTrips(t *testing.T) {
	if n, ok := exileSuspendCount(exileSuspendKind(3)); !ok || n != 3 {
		t.Errorf("round trip = %d, %v", n, ok)
	}
	for _, k := range []string{resolutionExitExileSuspend, "exile-suspend:x", "exile-suspend:0", resolutionExitWithCounter} {
		if _, ok := exileSuspendCount(k); ok {
			t.Errorf("%q read as an exile-suspend kind", k)
		}
	}
}
