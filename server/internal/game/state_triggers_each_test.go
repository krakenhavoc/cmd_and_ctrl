package game

import (
	"testing"

	"github.com/google/uuid"
)

// state_triggers_each_test.go — CR 603.8 state triggers about each of
// several objects (#1858, Bomb Squad): "Whenever a creature has four or
// more fuse counters on it, …". The state is a fact about one creature,
// so the ability triggers once for each creature in that state, and the
// latch is per creature as well as per source: an item about one
// creature waiting or on the stack does not stop the ability triggering
// for another.

const (
	stEachKey         = "Each Probe — remove a creature's charge counters"
	stEachStubbornKey = "Each Probe — you gain 1 life for a charged creature"
	stEachChargeState = "test/each-creature-two-or-more-charge-counters"
)

func init() {
	RegisterStateEachCondition(stEachChargeState, func(_ *Game, _ *Card, _ uuid.UUID, obj *Card) bool {
		return obj.IsCreature() && obj.Counters[CounterCharge] >= 2
	})
}

// stEachRow removes the charge counters from the creature the item is
// about, which ends that creature's state.
func stEachRow() TriggeredAbility {
	return TriggeredAbility{
		Key:   stEachKey,
		State: stEachChargeState,
		Effect: func(g *Game, item *StackItem) error {
			if item.Trigger == nil || item.Trigger.Object == nil {
				return nil
			}
			c := findBattlefieldCard(g, item.Trigger.Object.ID)
			if c == nil || c.ObjectEpoch != item.Trigger.Object.Epoch {
				return nil
			}
			return g.AddCounterForEffect(c.InstanceID, CounterCharge, -c.Counters[CounterCharge])
		},
	}
}

// stEachStubbornRow leaves the creature's state alone, so the ability
// triggers for it again each time the item leaves the stack.
func stEachStubbornRow() TriggeredAbility {
	return TriggeredAbility{
		Key:   stEachStubbornKey,
		State: stEachChargeState,
		Effect: func(g *Game, item *StackItem) error {
			return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, 1)
		},
	}
}

// stEachItems counts the items of `key` from `source` about `about`, in
// the queue and on the stack.
func stEachItems(g *Game, source, about uuid.UUID, key string) (queued, stacked int) {
	is := func(it *StackItem) bool {
		return it.SourceCardID == source && it.Label == key &&
			it.Trigger != nil && it.Trigger.Object != nil && it.Trigger.Object.ID == about
	}
	g.ReadSnapshot(func() {
		for _, it := range g.PendingTriggers {
			if is(it) {
				queued++
			}
		}
		for _, it := range g.StackMeta {
			if is(it) {
				stacked++
			}
		}
	})
	return queued, stacked
}

func pushEachCreature(t *testing.T, g *Game, owner *Player) uuid.UUID {
	t.Helper()
	id := uuid.New()
	pushTestCreature(g, id, owner, 2, 2)
	return id
}

// The ability triggers once for each creature in the state, and only
// once for each while that creature's item waits.
func TestStateEachTriggerTriggersOncePerObject(t *testing.T) {
	withStateCatalog(t, stEachRow())
	g := newActiveGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	probe := pushStateProbe(t, g, "Creature — Dwarf")
	a := pushEachCreature(t, g, me)
	b := pushEachCreature(t, g, bob)

	addCharge(t, g, a, 1)
	if q, s := stEachItems(g, probe, a, stEachKey); q+s != 0 {
		t.Fatalf("one counter triggered the ability (%d queued, %d stacked)", q, s)
	}
	addCharge(t, g, a, 1)
	if q, _ := stEachItems(g, probe, a, stEachKey); q != 1 {
		t.Fatalf("queued about A = %d after its second counter, want 1", q)
	}
	addCharge(t, g, a, 3)
	if q, _ := stEachItems(g, probe, a, stEachKey); q != 1 {
		t.Fatalf("queued about A = %d after more counters, want 1: the item about A latches it (CR 603.8)", q)
	}
	// B's state is a different state: A's waiting item does not latch it.
	addCharge(t, g, b, 2)
	if q, _ := stEachItems(g, probe, b, stEachKey); q != 1 {
		t.Fatalf("queued about B = %d, want 1: the latch names the creature", q)
	}
	g.WithWriteLock(func() { g.runStateChecksLocked() })
	_, sa := stEachItems(g, probe, a, stEachKey)
	_, sb := stEachItems(g, probe, b, stEachKey)
	if sa != 1 || sb != 1 {
		t.Fatalf("on the stack: %d about A, %d about B; want one each", sa, sb)
	}
	// Resolving both ends both states; nothing retriggers.
	g.WithWriteLock(func() {
		g.resolveTopAbilityLocked()
		g.resolveTopAbilityLocked()
		g.runStateChecksLocked()
	})
	for _, id := range []uuid.UUID{a, b} {
		if q, s := stEachItems(g, probe, id, stEachKey); q+s != 0 {
			t.Fatalf("after resolving: %d queued, %d stacked about %v; want none", q, s, id)
		}
		g.ReadSnapshot(func() {
			if c := g.findCardByIDLocked(id); c == nil || c.Counters[CounterCharge] != 0 {
				t.Fatalf("the item about %v did not act on its creature: %+v", id, c)
			}
		})
	}
}

// A permanent that is not a creature is not in the state the row reads,
// and the source's own state counts like any other creature's.
func TestStateEachTriggerReadsEveryPermanentThroughItsCondition(t *testing.T) {
	withStateCatalog(t, stEachRow())
	g := newActiveGame(t)
	me := g.Seats[0].ID
	probe := pushStateProbe(t, g, "Creature — Dwarf")
	rock := pushTypedTestCard(g, Card{Name: "Rock", TypeLine: "Artifact", Owner: me, Controller: me})
	addCharge(t, g, rock, 3)
	if q, s := stEachItems(g, probe, rock, stEachKey); q+s != 0 {
		t.Fatalf("an artifact triggered a creature state (%d queued, %d stacked)", q, s)
	}
	addCharge(t, g, probe, 2)
	if q, _ := stEachItems(g, probe, probe, stEachKey); q != 1 {
		t.Fatalf("queued about the source itself = %d, want 1", q)
	}
}

// CR 603.8: "Then, if the object with the ability is still in the same
// zone and the game state still matches its trigger condition, the
// ability will trigger again" — for the same creature.
func TestStateEachTriggerTriggersAgainForTheSameObject(t *testing.T) {
	withStateCatalog(t, stEachStubbornRow())
	g := newActiveGame(t)
	me := g.Seats[0]
	probe := pushStateProbe(t, g, "Creature — Dwarf")
	a := pushEachCreature(t, g, me)
	addCharge(t, g, a, 2)
	life := me.Life
	for round := 1; round <= 3; round++ {
		g.WithWriteLock(func() { g.runStateChecksLocked() })
		if q, s := stEachItems(g, probe, a, stEachStubbornKey); q != 0 || s != 1 {
			t.Fatalf("round %d: %d queued, %d stacked; want one on the stack", round, q, s)
		}
		g.WithWriteLock(func() { g.resolveTopAbilityLocked() })
		if q, s := stEachItems(g, probe, a, stEachStubbornKey); q+s != 0 {
			t.Fatalf("round %d: retriggered during its own resolution (%d queued, %d stacked)", round, q, s)
		}
		if me.Life != life+round {
			t.Fatalf("round %d: life %d, want %d", round, me.Life, life+round)
		}
	}
}

// CR 400.7: a creature that left and came back is a new object in a
// state of its own; the item about the old object does not latch it.
func TestStateEachTriggerLatchNamesTheObject(t *testing.T) {
	withStateCatalog(t, stEachStubbornRow())
	g := newActiveGame(t)
	me := g.Seats[0]
	probe := pushStateProbe(t, g, "Creature — Dwarf")
	a := pushEachCreature(t, g, me)
	addCharge(t, g, a, 2)
	if q, _ := stEachItems(g, probe, a, stEachStubbornKey); q != 1 {
		t.Fatalf("queued = %d, want 1", q)
	}
	g.WithWriteLock(func() {
		findBattlefieldCard(g, a).ObjectEpoch++ // as if it had left and come back
		g.stateTriggersLocked()
	})
	if q, _ := stEachItems(g, probe, a, stEachStubbornKey); q != 2 {
		t.Fatalf("queued = %d, want 2: the new object is not latched by the old object's item", q)
	}
}

// Two sources with the ability each trigger for the same creature
// (Bomb Squad's ruling: two Bomb Squads both trigger).
func TestStateEachTriggerLatchIsPerSource(t *testing.T) {
	withStateCatalog(t, stEachRow())
	g := newActiveGame(t)
	me := g.Seats[0]
	p1 := pushStateProbe(t, g, "Creature — Dwarf")
	p2 := pushStateProbe(t, g, "Creature — Dwarf")
	a := pushEachCreature(t, g, me)
	addCharge(t, g, a, 2)
	q1, _ := stEachItems(g, p1, a, stEachKey)
	q2, _ := stEachItems(g, p2, a, stEachKey)
	if q1 != 1 || q2 != 1 {
		t.Fatalf("queued: %d from the first source, %d from the second; want one each", q1, q2)
	}
}

// An instance whose "you may" is still open latches the ability for its
// own creature only: the open prompt carries the creature too.
func TestStateEachTriggerOpenPromptLatchesItsObjectOnly(t *testing.T) {
	row := stEachRow()
	row.OptionalPrompt = &TriggerOptionalPrompt{Question: "Remove the counters?"}
	withStateCatalog(t, row)
	g := newActiveGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	pushStateProbe(t, g, "Creature — Dwarf")
	a := pushEachCreature(t, g, me)
	b := pushEachCreature(t, g, bob)
	prompts := func(about uuid.UUID) (n int) {
		g.ReadSnapshot(func() {
			for _, c := range g.PendingChoices {
				if c != nil && c.triggerResume != nil && c.triggerResume.ability.Key == stEachKey &&
					c.triggerResume.tc.Object != nil && c.triggerResume.tc.Object.ID == about {
					n++
				}
			}
		})
		return n
	}
	addCharge(t, g, a, 2)
	addCharge(t, g, a, 1)
	if n := prompts(a); n != 1 {
		t.Fatalf("%d prompts about A, want 1: the open prompt latches the ability for A", n)
	}
	addCharge(t, g, b, 2)
	if n := prompts(b); n != 1 {
		t.Fatalf("%d prompts about B, want 1: A's open prompt does not latch B", n)
	}
}

// The latch lives on the item's trigger context, which is snapshot data:
// a restored table's waiting item still latches the ability for its
// creature, and still resolves against that creature.
func TestStateEachTriggerLatchSurvivesARestore(t *testing.T) {
	withStateCatalog(t, stEachRow())
	g := newActiveGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	probe := pushStateProbe(t, g, "Creature — Dwarf")
	a := pushEachCreature(t, g, me)
	b := pushEachCreature(t, g, bob)
	addCharge(t, g, a, 2)
	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("a waiting per-object state trigger blocks the restore point: %+v", snap.Continuations)
	}
	restored, err := throughJSON(t, snap).RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	restored.WithWriteLock(func() { restored.stateTriggersLocked() })
	if q, _ := stEachItems(restored, probe, a, stEachKey); q != 1 {
		t.Fatalf("queued about A after restore = %d, want 1: the restored item still latches it", q)
	}
	addCharge(t, restored, b, 2)
	if q, _ := stEachItems(restored, probe, b, stEachKey); q != 1 {
		t.Fatalf("queued about B after restore = %d, want 1", q)
	}
	restored.WithWriteLock(func() {
		restored.runStateChecksLocked()
		restored.resolveTopAbilityLocked()
		restored.resolveTopAbilityLocked()
	})
	restored.ReadSnapshot(func() {
		for _, id := range []uuid.UUID{a, b} {
			if c := restored.findCardByIDLocked(id); c == nil || c.Counters[CounterCharge] != 0 {
				t.Fatalf("the restored item about %v did not resolve against it: %+v", id, c)
			}
		}
	})
}
