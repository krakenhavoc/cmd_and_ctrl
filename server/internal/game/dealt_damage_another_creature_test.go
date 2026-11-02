package game

import (
	"testing"

	"github.com/google/uuid"
)

// dealt_damage_another_creature_test.go — #4460, Wolverine, Best There
// Is: "if Wolverine dealt damage to another creature this turn". The
// damage record marks a victim that was a creature as it was dealt, and
// ObjectDealtDamageToAnotherCreatureThisTurn reads it for one source
// object.

func dealtDamageToAnotherCreature(t *testing.T, g *Game, source uuid.UUID) bool {
	t.Helper()
	ref := liveObjectRef(t, g, source)
	var got bool
	g.WithWriteLock(func() { got = g.ObjectDealtDamageToAnotherCreatureThisTurn(ref) })
	return got
}

func TestDealtDamageToAnotherCreatureCountsOnlyCreatureVictims(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	source := pushScopedTestCreature(g, me, 3, 3)
	other := pushScopedTestCreature(g, me, 3, 3)
	walker := uuid.New()
	g.Battlefield.PushTop(Card{InstanceID: walker, Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		Owner: opp, Controller: opp, Counters: map[string]int{CounterLoyalty: 5}})
	victim := pushScopedTestCreature(g, opp, 2, 6)

	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(source, opp, 2); err != nil {
			t.Fatal(err)
		}
		if err := g.DealDamageToCreatureForEffect(source, walker, 1); err != nil {
			t.Fatal(err)
		}
		if err := g.DealDamageToCreatureForEffect(other, victim, 1); err != nil {
			t.Fatal(err)
		}
	})
	if dealtDamageToAnotherCreature(t, g, source) {
		t.Fatal("damage to a player and a planeswalker read as damage to a creature")
	}
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(source, victim, 1); err != nil {
			t.Fatal(err)
		}
	})
	if !dealtDamageToAnotherCreature(t, g, source) {
		t.Error("damage to a creature is not read")
	}
	if dealtDamageToAnotherCreature(t, g, victim) {
		t.Error("the read is per source")
	}
	var recs []ObjectDamageRecord
	g.ReadSnapshot(func() { recs = g.TurnTally.DamagedObjects })
	for _, r := range recs {
		if r.Victim.ID == walker && r.VictimCreature {
			t.Error("the planeswalker's record is marked a creature")
		}
	}
}

// Damage the source deals to itself is not "another creature".
func TestDealtDamageToAnotherCreatureExcludesTheSourceItself(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	source := pushScopedTestCreature(g, me, 3, 5)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(source, source, 1); err != nil {
			t.Fatal(err)
		}
	})
	if dealtDamageToAnotherCreature(t, g, source) {
		t.Error("damage to itself read as damage to another creature")
	}
}
