package game

import (
	"testing"

	"github.com/google/uuid"
)

// exile_with_counter_test.go — "exile that card with a dream counter on
// it instead of putting it into your graveyard as it resolves" (Goliath
// Daydreamer, #3213, ADR 0107 amendment of 2026-10-13).

// castMarked casts a plain sorcery from hand and marks it.
func castMarked(t *testing.T, g *Game, me *Player) uuid.UUID {
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
	g.WithWriteLock(func() { ok = g.MarkSpellExilesWithCounterForEffect(c.InstanceID, CounterDream) })
	if !ok {
		t.Fatal("the spell on the stack could not be marked")
	}
	return c.InstanceID
}

func TestMarkedSpellIsExiledWithACounterWhenItResolves(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := castMarked(t, g, me)
	resolveTop(t, g)
	if me.Graveyard.Contains(id) {
		t.Error("a marked spell that resolved went to the graveyard")
	}
	c := exiledCardByIDLocked(g, id)
	if c == nil {
		t.Fatal("a marked spell that resolved was not exiled")
	}
	if c.Counters[CounterDream] != 1 {
		t.Errorf("counters = %v, want one dream counter", c.Counters)
	}
}

func TestMarkedSpellThatIsCounteredGoesToTheGraveyard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := castMarked(t, g, me)
	if err := g.CounterSpell(id, nil); err != nil {
		t.Fatalf("CounterSpell: %v", err)
	}
	if !me.Graveyard.Contains(id) || g.Exile.Contains(id) {
		t.Error("a countered marked spell must go to the graveyard")
	}
}

func TestMarkingRefusesAnythingButASpellOnTheStack(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	var ok bool
	g.WithWriteLock(func() { ok = g.MarkSpellExilesWithCounterForEffect(uuid.New(), CounterDream) })
	if ok {
		t.Error("an unknown ID was marked")
	}
	id := castMarked(t, g, me)
	g.WithWriteLock(func() { ok = g.MarkSpellExilesWithCounterForEffect(id, "") })
	if ok {
		t.Error("an empty counter kind was accepted")
	}
}

// CR 616.1: a marked rebound spell has two replacements for one event,
// so the controller chooses; picking the dream exile leaves no rebound
// trigger behind.
func TestMarkedReboundSpellAsksTheControllerWhichExitApplies(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	c := seedReboundCard(me, "Test Rebound", "Instant")
	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	var ok bool
	g.WithWriteLock(func() { ok = g.MarkSpellExilesWithCounterForEffect(c.InstanceID, CounterDream) })
	if !ok {
		t.Fatal("could not mark the rebound spell")
	}
	resolveTop(t, g)
	var pick *PendingChoice
	for _, pc := range g.PendingChoices {
		if pc != nil && pc.Kind == PendingChoiceOptionPick {
			pick = pc
		}
	}
	if pick == nil {
		t.Fatal("two applicable exits should have asked the controller")
	}
	// Rebound is offered first, the dream exile second.
	if err := g.ResolveOptionPick(pick.ID, me.ID, 1); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	card := exiledCardByIDLocked(g, c.InstanceID)
	if card == nil || card.Counters[CounterDream] != 1 {
		t.Fatalf("exiled card = %+v, want a dream counter", card)
	}
	if n := len(reboundTriggers(g)); n != 0 {
		t.Errorf("the dream exile was chosen but %d rebound triggers were created", n)
	}
}
