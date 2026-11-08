package game

import (
	"testing"

	"github.com/google/uuid"
)

// one_more_counter_test.go — ModOneMoreCounter (#4973, Brad Boimler):
// one more counter of each kind on the record's player's permanents,
// until cleanup.

func pushOneMoreCounterBear(g *Game, controller *Player) uuid.UUID {
	c := NewCard("Bear", controller.ID)
	c.TypeLine = "Creature — Bear"
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

func TestOneMoreCounterAddsOneOfEachKindOnYourPermanentsThisTurn(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	var mine, theirs uuid.UUID
	g.WithWriteLock(func() {
		mine = pushOneMoreCounterBear(g, me)
		theirs = pushOneMoreCounterBear(g, opp)
		if !g.OneMoreCounterThisTurnForEffect(uuid.Nil, me.ID, "Brad Boimler") {
			t.Fatal("no record was written")
		}
		_ = g.AddCounterForEffect(mine, CounterPlusOne, 2)
		_ = g.AddCounterForEffect(mine, "charge", 1)
		_ = g.AddCounterForEffect(theirs, CounterPlusOne, 1)
		_ = g.AddPlayerCounterForEffect(me.ID, "poison", 1)
	})
	read := func(id uuid.UUID, kind string) int {
		c, _ := g.LookupCardForEffect(id)
		return c.Counters[kind]
	}
	if got := read(mine, CounterPlusOne); got != 3 {
		t.Errorf("+1/+1 counters on mine = %d, want 3", got)
	}
	if got := read(mine, "charge"); got != 2 {
		t.Errorf("charge counters on mine = %d, want 2", got)
	}
	if got := read(theirs, CounterPlusOne); got != 1 {
		t.Errorf("an opponent's permanent got %d, want 1", got)
	}
	if got := me.Counters["poison"]; got != 1 {
		t.Errorf("a player got %d poison counters, want 1 (a player is not a permanent)", got)
	}
	g.WithWriteLock(func() {
		g.sweepScopedEffectsLocked(true)
		_ = g.AddCounterForEffect(mine, CounterPlusOne, 1)
	})
	if got := read(mine, CounterPlusOne); got != 4 {
		t.Errorf("after cleanup: +1/+1 counters = %d, want 4", got)
	}
}

func TestOneMoreCounterRefusesNoPlayer(t *testing.T) {
	g := newActiveGame(t)
	g.WithWriteLock(func() {
		if g.OneMoreCounterThisTurnForEffect(uuid.Nil, uuid.Nil, "x") {
			t.Error("a record with no player was written")
		}
	})
}

// ModOneMorePlusOneCounter (#5235, Prairie Dog): one more +1/+1 counter,
// only when the record's player puts +1/+1 counters on a creature they
// control, until cleanup.
func TestOneMorePlusOneCounterAddsOneOnlyWhenYouPutThemOnYourCreature(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	var mine, theirs, rock uuid.UUID
	g.WithWriteLock(func() {
		mine = pushOneMoreCounterBear(g, me)
		theirs = pushOneMoreCounterBear(g, opp)
		r := NewCard("Rock", me.ID)
		r.TypeLine = "Artifact"
		g.Battlefield.PushTop(r)
		rock = r.InstanceID
		if !g.OneMorePlusOneCounterThisTurnForEffect(uuid.Nil, me.ID, "Prairie Dog") {
			t.Fatal("no record was written")
		}
		_ = g.AddCounterForEffect(mine, CounterPlusOne, 2)
		_ = g.AddCounterForEffect(mine, "charge", 1)
		_ = g.AddCounterForEffect(rock, CounterPlusOne, 1)
		_ = g.AddCounterForEffect(theirs, CounterPlusOne, 1)
		// An opponent putting +1/+1 counters on my creature is not "you
		// would put".
		pop := g.pushCounterPlacerLocked(opp.ID)
		_ = g.AddCounterForEffect(mine, CounterPlusOne, 1)
		pop()
	})
	read := func(id uuid.UUID, kind string) int {
		c, _ := g.LookupCardForEffect(id)
		return c.Counters[kind]
	}
	if got := read(mine, CounterPlusOne); got != 4 {
		t.Errorf("+1/+1 counters on mine = %d, want 3 from my two and 1 an opponent put", got)
	}
	if got := read(mine, "charge"); got != 1 {
		t.Errorf("charge counters on mine = %d, want 1 (only +1/+1 counters)", got)
	}
	if got := read(rock, CounterPlusOne); got != 1 {
		t.Errorf("a noncreature permanent got %d, want 1", got)
	}
	if got := read(theirs, CounterPlusOne); got != 1 {
		t.Errorf("an opponent's creature got %d, want 1", got)
	}
	g.WithWriteLock(func() {
		g.sweepScopedEffectsLocked(true)
		_ = g.AddCounterForEffect(mine, CounterPlusOne, 1)
	})
	if got := read(mine, CounterPlusOne); got != 5 {
		t.Errorf("after cleanup: +1/+1 counters = %d, want 5", got)
	}
}

func TestOneMorePlusOneCounterRefusesNoPlayer(t *testing.T) {
	g := newActiveGame(t)
	g.WithWriteLock(func() {
		if g.OneMorePlusOneCounterThisTurnForEffect(uuid.Nil, uuid.Nil, "x") {
			t.Error("a record with no player was written")
		}
	})
}
