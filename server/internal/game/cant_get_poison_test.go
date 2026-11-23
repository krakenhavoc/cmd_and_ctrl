package game

import (
	"testing"

	"github.com/google/uuid"
)

// cant_get_poison_test.go — Melira, the Living Cure's "you can't get
// additional poison counters this turn" (CR 101.2), asked before the
// counter window opens.

func TestCantGetPoisonStopsPoisonOnlyForThatPlayer(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		if !g.PlayerCantGetPoisonThisTurnForEffect(uuid.Nil, me.ID, "Melira") {
			t.Fatal("no record written")
		}
		if err := g.AddPlayerCounterForEffect(me.ID, CounterPoison, 2); err != nil {
			t.Fatal(err)
		}
		if err := g.AddPlayerCounterForEffect(me.ID, CounterEnergy, 2); err != nil {
			t.Fatal(err)
		}
		if err := g.AddPlayerCounterForEffect(them.ID, CounterPoison, 2); err != nil {
			t.Fatal(err)
		}
	})
	if got := me.Counters[CounterPoison]; got != 0 {
		t.Errorf("poison %d, want 0 for the banned player", got)
	}
	if got := me.Counters[CounterEnergy]; got != 2 {
		t.Errorf("energy %d, want 2: only poison is stopped", got)
	}
	if got := them.Counters[CounterPoison]; got != 2 {
		t.Errorf("another player's poison %d, want 2", got)
	}
}

func TestCantGetPoisonRegistersNothingForAnUnseatedPlayer(t *testing.T) {
	g := newActiveGame(t)
	g.WithWriteLock(func() {
		if g.PlayerCantGetPoisonThisTurnForEffect(uuid.Nil, uuid.New(), "Melira") {
			t.Error("a record was written for a player who is not seated")
		}
	})
}
