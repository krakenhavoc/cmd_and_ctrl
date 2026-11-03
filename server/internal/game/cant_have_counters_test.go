package game

import (
	"testing"

	"github.com/google/uuid"
)

// cant_have_counters_test.go — #4692's engine half: "can't have counters
// put on it" stops every placement before the CR 614 window (so no
// replacement sees it), stops wither's -1/-1 counters, makes a counter
// cost unpayable, keeps the permanent out of blight and proliferate, and
// leaves removal alone. The Aura form is in cards/effects
// (blossombind_test.go).

const testCantHaveCountersOracle = "test-cant-have-counters"

func withCantHaveCountersStatic(t *testing.T) {
	t.Helper()
	prev := CatalogStaticAbilities
	t.Cleanup(func() { CatalogStaticAbilities = prev })
	CatalogStaticAbilities = func(key string) []StaticAbility {
		if key == testCantHaveCountersOracle {
			return []StaticAbility{{
				Layer:     Layer6Ability,
				AppliesTo: func(target *Card, _ *Game, source *Card) bool { return target.InstanceID == source.InstanceID },
				Apply:     func(c *Characteristic, _ *Card, _ *Game, _ *Card) { c.CantHaveCounters = true },
			}}
		}
		if prev == nil {
			return nil
		}
		return prev(key)
	}
}

// pushCounterlessCreature puts a creature that can't have counters put on
// it under `owner`, starting with `plus` +1/+1 counters. Caller holds g.mu.
func pushCounterlessCreature(g *Game, owner uuid.UUID, plus int) uuid.UUID {
	id := uuid.New()
	pushPrintedKeywordCreature(g, id, owner, owner, 2, 2)
	c := findBattlefieldCard(g, id)
	c.OracleID = testCantHaveCountersOracle
	if plus > 0 {
		c.Counters = map[string]int{CounterPlusOne: plus}
	}
	g.layerVersion.Add(1)
	g.RecomputeLayersIfStaleLocked()
	return id
}

func countersOn(g *Game, id uuid.UUID, kind string) int {
	c := findBattlefieldCard(g, id)
	if c == nil {
		return -1
	}
	return c.Counters[kind]
}

func TestCantHaveCountersRefusesPlacementBeforeTheWindow(t *testing.T) {
	withCantHaveCountersStatic(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	rec := &counterEventRecorder{}
	var bound uuid.UUID
	g.WithWriteLock(func() {
		bound = pushCounterlessCreature(g, me.ID, 1)
		g.RegisterReplacementForTest(rec.effect())
		before := len(iwtEventsOfKind(g, EventCounterPlaced))
		for _, kind := range []string{CounterPlusOne, CounterMinusOne, "shield"} {
			if err := g.AddCounterForEffect(bound, kind, 2); err != nil {
				t.Fatalf("AddCounterForEffect(%s): %v", kind, err)
			}
		}
		if n, err := g.AddCounterMustSettleNowForEffect(bound, CounterPlusOne, 1); err != nil || n != 0 {
			t.Errorf("AddCounterMustSettleNowForEffect = %d, %v; want 0, nil", n, err)
		}
		if n := len(iwtEventsOfKind(g, EventCounterPlaced)) - before; n != 0 {
			t.Errorf("%d counter-placed events for refused placements", n)
		}
	})
	if got := countersOn(g, bound, CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters = %d, want the 1 it already had", got)
	}
	if got := countersOn(g, bound, CounterMinusOne) + countersOn(g, bound, "shield"); got != 0 {
		t.Errorf("%d -1/-1 or shield counters landed", got)
	}
	if len(rec.seen) != 0 {
		t.Errorf("the counter window saw %d refused placements; a replacement must never see them", len(rec.seen))
	}
	// The sandbox edit reaches the counter door directly.
	if err := g.AddCounter(bound, CounterPlusOne, 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if got := countersOn(g, bound, CounterPlusOne); got != 1 {
		t.Errorf("a sandbox edit put a counter on it: %d", got)
	}
	// Removing is not putting.
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(bound, CounterPlusOne, -1); err != nil {
			t.Fatalf("remove: %v", err)
		}
	})
	if got := countersOn(g, bound, CounterPlusOne); got != 0 {
		t.Errorf("+1/+1 counters after a removal = %d, want 0", got)
	}
}

func TestCantHaveCountersWitherDamageMarksNothing(t *testing.T) {
	withCantHaveCountersStatic(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := uuid.New()
	rec := &counterEventRecorder{}
	var bound uuid.UUID
	g.WithWriteLock(func() {
		pushPrintedKeywordCreature(g, src, me.ID, me.ID, 3, 3, KeywordWither)
		bound = pushCounterlessCreature(g, opp.ID, 0)
		g.RegisterReplacementForTest(rec.effect())
		if err := g.DealDamageToCreatureForEffect(src, bound, 3); err != nil {
			t.Fatalf("DealDamageToCreatureForEffect: %v", err)
		}
	})
	if got := minusOneCountersOn(g, bound); got != 0 {
		t.Errorf("-1/-1 counters = %d, want 0", got)
	}
	if c := findBattlefieldCard(g, bound); c == nil || c.DamageMarked != 0 {
		t.Errorf("wither damage was marked or the creature died: %+v", c)
	}
	if len(rec.seen) != 0 {
		t.Error("the counter window saw the refused wither counters")
	}
}

func TestCantHaveCountersCostsBlightAndProliferate(t *testing.T) {
	withCantHaveCountersStatic(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		bound := pushCounterlessCreature(g, me.ID, 1)
		other := uuid.New()
		pushPrintedKeywordCreature(g, other, me.ID, me.ID, 2, 2)
		findBattlefieldCard(g, other).Counters = map[string]int{CounterPlusOne: 1}

		if g.canPlaceCounterLocked(me.ID, bound, &CounterAddCost{Counter: CounterMinusOne, N: 1}) {
			t.Error("a counter cost on it can be paid (CR 118.3)")
		}
		if !g.canPlaceCounterLocked(me.ID, other, &CounterAddCost{Counter: CounterMinusOne, N: 1}) {
			t.Error("a counter cost on an unrestricted creature can't be paid")
		}
		for _, id := range g.BlightOptionsForEffect(me.ID) {
			if id == bound {
				t.Error("it is a blight option (CR 701.68b)")
			}
		}
		if err := g.validateBlightCreatureLocked(me.ID, bound); err == nil {
			t.Error("a blight naming it was accepted")
		}
		cards, _ := g.ProliferateCandidatesForEffect()
		offered := map[uuid.UUID]bool{}
		for _, id := range cards {
			offered[id] = true
		}
		if offered[bound] || !offered[other] {
			t.Errorf("proliferate offered it %v, the other creature %v", offered[bound], offered[other])
		}
		suggested, _ := g.ProliferateSuggestionForEffect(me.ID)
		for _, id := range suggested {
			if id == bound {
				t.Error("proliferate suggests it")
			}
		}
	})
}
