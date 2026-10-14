package game

import (
	"testing"

	"github.com/google/uuid"
)

// shield_counters_test.go — CR 122.1c (#2000): a shield counter's
// replacement effect (destruction by an effect) and prevention effect
// (damage), shield_counters.go.

// shieldedBear puts a 2/2 with n shield counters onto the battlefield.
func shieldedBear(t *testing.T, g *Game, owner *Player, n int) uuid.UUID {
	t.Helper()
	id := pushColouredCreature(g, owner, "Shielded Bear", []string{"G"})
	g.WithWriteLock(func() {
		if err := g.applyCounterLocked(id, CounterShield, n); err != nil {
			t.Fatal(err)
		}
	})
	return id
}

func shieldCountersOn(g *Game, id uuid.UUID) int {
	if c := findBattlefieldCard(g, id); c != nil {
		return c.Counters[CounterShield]
	}
	return -1
}

// "If this permanent would be destroyed as the result of an effect,
// instead remove a shield counter from it." Two counters are two
// destructions, with no ordering question between them.
func TestShieldCounterReplacesDestructionByAnEffect(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := shieldedBear(t, g, me, 2)

	destroy(t, g, id)
	if got := shieldCountersOn(g, id); got != 1 {
		t.Fatalf("after one destruction: %d shield counters (-1 is gone), want 1", got)
	}
	if len(g.PendingChoices) != 0 {
		t.Fatalf("two shield counters asked a question: %+v", g.PendingChoices[0])
	}
	destroy(t, g, id)
	if got := shieldCountersOn(g, id); got != 0 {
		t.Fatalf("after two destructions: %d shield counters, want 0", got)
	}
	destroy(t, g, id)
	if findBattlefieldCard(g, id) != nil || !me.Graveyard.Contains(id) {
		t.Error("with no shield counter left, the third destruction kills it")
	}
}

// Damage that can't be prevented is dealt, and the prevention effect's
// additional effect still happens, so a counter is removed (CR 615.12).
// The lethal damage then destroys the creature by a state-based action,
// which is not "the result of an effect", so the second counter doesn't
// replace it.
func TestShieldCounterDoesNotStopLethalDamageThatCantBePrevented(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := shieldedBear(t, g, opp, 2)
	bolt := pushRedSpell(g, me)

	g.WithWriteLock(func() {
		if err := g.DealMarkedDamageForEffect(bolt, nil, id, 2, DamageMarks{CantBePrevented: true}); err != nil {
			t.Fatal(err)
		}
	})
	if got := damageOn(g, id); got != 2 {
		t.Fatalf("damage that can't be prevented: %d marked, want 2", got)
	}
	if got := shieldCountersOn(g, id); got != 1 {
		t.Fatalf("after unpreventable damage: %d shield counters, want 1 (CR 615.12 still removes one)", got)
	}
	g.RunStateChecksForTest()
	if findBattlefieldCard(g, id) != nil {
		t.Error("lethal damage is a state-based destruction; the shield counter doesn't replace it")
	}
}

// "If damage would be dealt to this permanent, prevent that damage and
// remove a shield counter from it."
func TestShieldCounterPreventsDamage(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := shieldedBear(t, g, opp, 1)
	bolt := pushRedSpell(g, me)

	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(bolt, id, 3); err != nil {
			t.Fatal(err)
		}
	})
	if got := damageOn(g, id); got != 0 {
		t.Errorf("shielded creature has %d damage, want 0", got)
	}
	if got := shieldCountersOn(g, id); got != 0 {
		t.Errorf("%d shield counters after preventing damage, want 0", got)
	}
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(bolt, id, 1); err != nil {
			t.Fatal(err)
		}
	})
	if got := damageOn(g, id); got != 1 {
		t.Errorf("with the counter gone the next damage is dealt: %d marked, want 1", got)
	}
}

// Damage dealt at the same time by two sources is all prevented for one
// counter; a second instruction is a second instance and spends another.
func TestShieldCounterSpendsOneCounterPerInstanceOfDamage(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := shieldedBear(t, g, opp, 2)
	first, second := pushRedSpell(g, me), pushRedSpell(g, me)

	g.WithWriteLock(func() {
		if err := g.DamageInstanceForEffect(func() error {
			if err := g.DealDamageToCreatureForEffect(first, id, 2); err != nil {
				return err
			}
			return g.DealDamageToCreatureForEffect(second, id, 2)
		}); err != nil {
			t.Fatal(err)
		}
	})
	if got := damageOn(g, id); got != 0 {
		t.Fatalf("simultaneous damage from two sources: %d marked, want 0", got)
	}
	if got := shieldCountersOn(g, id); got != 1 {
		t.Fatalf("one instance of damage spent %d counters, want 1", 2-got)
	}
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(first, id, 2); err != nil {
			t.Fatal(err)
		}
	})
	if got := shieldCountersOn(g, id); got != 0 || damageOn(g, id) != 0 {
		t.Errorf("a second instance: %d counters, %d damage, want 0 and 0", got, damageOn(g, id))
	}
}
