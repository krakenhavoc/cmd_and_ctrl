package game

import (
	"testing"

	"github.com/google/uuid"
)

// state_triggers_test.go — CR 603.8 state triggers (ADR 0107 §1,
// #1858). A row that declares State watches no event: the engine asks
// it after every event and in each pass of the CR 704.3 loop, and the
// ability is latched while an item of it from the same object is
// waiting, being announced, on the stack or resolving.

const stOracle = "fixture-state-trigger"

const (
	stThresholdKey = "State Probe — remove the charge counters"
	stStubbornKey  = "State Probe — you gain 1 life"
	stCrocKey      = "State Croc — sacrifice it"
)

// stThresholdRow is "When there are two or more charge counters on this
// permanent, remove them": resolving it ends the state.
func stThresholdRow() TriggeredAbility {
	return TriggeredAbility{
		Key: stThresholdKey,
		State: func(_ *Game, source *Card, _ uuid.UUID) bool {
			return source.Counters[CounterCharge] >= 2
		},
		Effect: func(g *Game, item *StackItem) error {
			c := g.findCardByIDLocked(item.SourceCardID)
			if c == nil {
				return nil
			}
			return g.AddCounterForEffect(item.SourceCardID, CounterCharge, -c.Counters[CounterCharge])
		},
	}
}

// stStubbornRow is the same condition with an effect that leaves the
// state alone, so it triggers again each time it leaves the stack.
func stStubbornRow() TriggeredAbility {
	return TriggeredAbility{
		Key: stStubbornKey,
		State: func(_ *Game, source *Card, _ uuid.UUID) bool {
			return source.Counters[CounterCharge] >= 2
		},
		Effect: func(g *Game, item *StackItem) error {
			return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, 1)
		},
	}
}

// stCrocRow is Emperor Crocodile's "When you control no other
// creatures, sacrifice this creature".
func stCrocRow() TriggeredAbility {
	return TriggeredAbility{
		Key: stCrocKey,
		State: func(g *Game, source *Card, controller uuid.UUID) bool {
			for i := range g.Battlefield.Cards {
				c := &g.Battlefield.Cards[i]
				if c.InstanceID != source.InstanceID && c.Controller == controller && c.IsCreature() {
					return false
				}
			}
			return true
		},
		Effect: func(g *Game, item *StackItem) error {
			return g.SacrificePermanentForEffect(item.SourceCardID)
		},
	}
}

func withStateCatalog(t *testing.T, rows ...TriggeredAbility) {
	t.Helper()
	d := &CardDef{Triggered: rows}
	IdentifyCatalogRows(stOracle, d)
	withCatalog(t, stOracle, d)
}

func pushStateProbe(t *testing.T, g *Game, typeLine string) uuid.UUID {
	t.Helper()
	me := g.Seats[0].ID
	return pushTypedTestCard(g, Card{
		Name: "State Probe", TypeLine: typeLine, OracleID: stOracle,
		Owner: me, Controller: me, Power: 5, Toughness: 5,
	})
}

// stItems counts the items of `key` from `source` in the queue and on
// the stack.
func stItems(g *Game, source uuid.UUID, key string) (queued, stacked int) {
	g.ReadSnapshot(func() {
		for _, it := range g.PendingTriggers {
			if it.SourceCardID == source && it.Label == key {
				queued++
			}
		}
		for _, it := range g.StackMeta {
			if it.SourceCardID == source && it.Label == key {
				stacked++
			}
		}
	})
	return queued, stacked
}

func addCharge(t *testing.T, g *Game, id uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(id, CounterCharge, n); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})
}

// CR 603.8: the ability triggers as soon as the state matches, on the
// event that made it match — and only once while it waits, however many
// more events keep the state true.
func TestStateTriggerTriggersOnceWhenTheStateArises(t *testing.T) {
	withStateCatalog(t, stThresholdRow())
	g := newActiveGame(t)
	probe := pushStateProbe(t, g, "Enchantment")

	addCharge(t, g, probe, 1)
	if q, s := stItems(g, probe, stThresholdKey); q+s != 0 {
		t.Fatalf("one charge counter triggered the ability (%d queued, %d stacked)", q, s)
	}
	addCharge(t, g, probe, 1)
	if q, _ := stItems(g, probe, stThresholdKey); q != 1 {
		t.Fatalf("queued = %d after the second counter, want 1: the state arose on that event", q)
	}
	addCharge(t, g, probe, 1)
	addCharge(t, g, probe, 1)
	if q, _ := stItems(g, probe, stThresholdKey); q != 1 {
		t.Fatalf("queued = %d after two more counters, want 1: a waiting state trigger is latched (CR 603.8)", q)
	}
	// Onto the stack: still latched.
	g.WithWriteLock(func() { g.runStateChecksLocked() })
	if q, s := stItems(g, probe, stThresholdKey); q != 0 || s != 1 {
		t.Fatalf("after the drain: %d queued, %d stacked; want the one item on the stack", q, s)
	}
	addCharge(t, g, probe, 1)
	if q, s := stItems(g, probe, stThresholdKey); q != 0 || s != 1 {
		t.Fatalf("a counter while it is on the stack retriggered it: %d queued, %d stacked", q, s)
	}
	// Resolving it ends the state, so nothing triggers afterwards.
	g.WithWriteLock(func() {
		g.resolveTopAbilityLocked()
		g.runStateChecksLocked()
	})
	if q, s := stItems(g, probe, stThresholdKey); q+s != 0 {
		t.Fatalf("after resolving: %d queued, %d stacked; want none", q, s)
	}
	g.ReadSnapshot(func() {
		if c := g.findCardByIDLocked(probe); c == nil || c.Counters[CounterCharge] != 0 {
			t.Fatalf("the resolved ability did not remove the counters: %+v", c)
		}
	})
}

// CR 603.8: "Then, if the object with the ability is still in the same
// zone and the game state still matches its trigger condition, the
// ability will trigger again."
func TestStateTriggerTriggersAgainOnceItHasResolved(t *testing.T) {
	withStateCatalog(t, stStubbornRow())
	g := newActiveGame(t)
	probe := pushStateProbe(t, g, "Enchantment")
	addCharge(t, g, probe, 2)
	me := g.Seats[0]
	life := me.Life
	for round := 1; round <= 3; round++ {
		g.WithWriteLock(func() { g.runStateChecksLocked() })
		if q, s := stItems(g, probe, stStubbornKey); q != 0 || s != 1 {
			t.Fatalf("round %d: %d queued, %d stacked; want exactly one on the stack", round, q, s)
		}
		g.WithWriteLock(func() { g.resolveTopAbilityLocked() })
		if q, s := stItems(g, probe, stStubbornKey); q+s != 0 {
			t.Fatalf("round %d: the ability retriggered during its own resolution (%d queued, %d stacked)", round, q, s)
		}
		if me.Life != life+round {
			t.Fatalf("round %d: life %d, want %d", round, me.Life, life+round)
		}
	}
}

// CR 603.8's own example: a state that holds only for a moment in the
// middle of a resolution still triggers (owner decision 1: the check
// runs after every event, not only when a player would get priority).
func TestStateTriggerSeesAMomentaryState(t *testing.T) {
	withStateCatalog(t, stThresholdRow())
	g := newActiveGame(t)
	probe := pushStateProbe(t, g, "Enchantment")
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(probe, CounterCharge, 2); err != nil {
			t.Fatal(err)
		}
		if err := g.AddCounterForEffect(probe, CounterCharge, -2); err != nil {
			t.Fatal(err)
		}
	})
	if q, _ := stItems(g, probe, stThresholdKey); q != 1 {
		t.Fatalf("queued = %d: two counters that came and went inside one mutation must still trigger the ability", q)
	}
}

// The CR 704.3 pass is the authoritative check: a state that no event
// announced — here, a permanent that was put on the battlefield in that
// state without the engine seeing it arise — triggers when a player
// would next receive priority.
func TestStateTriggerIsAskedInTheStateBasedActionLoop(t *testing.T) {
	withStateCatalog(t, stThresholdRow())
	g := newActiveGame(t)
	me := g.Seats[0].ID
	probe := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: probe, Name: "State Probe", TypeLine: "Enchantment", OracleID: stOracle,
		Owner: me, Controller: me, Counters: map[string]int{CounterCharge: 3},
	})
	if q, s := stItems(g, probe, stThresholdKey); q+s != 0 {
		t.Fatalf("nothing has asked yet, but %d queued / %d stacked", q, s)
	}
	g.WithWriteLock(func() { g.runStateChecksLocked() })
	if _, s := stItems(g, probe, stThresholdKey); s != 1 {
		t.Fatalf("stacked = %d after the CR 704.3 loop, want 1", s)
	}
}

// CR 400.7: a permanent that changed zones is a new object, and its
// ability is a new ability — the old object's waiting item does not
// latch it.
func TestStateTriggerLatchIsPerObject(t *testing.T) {
	withStateCatalog(t, stStubbornRow())
	g := newActiveGame(t)
	probe := pushStateProbe(t, g, "Enchantment")
	addCharge(t, g, probe, 2)
	if q, _ := stItems(g, probe, stStubbornKey); q != 1 {
		t.Fatalf("queued = %d, want 1", q)
	}
	g.WithWriteLock(func() {
		c := g.findCardByIDLocked(probe)
		c.ObjectEpoch++ // as if it had left and come back
		g.stateTriggersLocked()
	})
	if q, _ := stItems(g, probe, stStubbornKey); q != 2 {
		t.Fatalf("queued = %d, want 2: the new object's ability is not latched by the old object's item", q)
	}
}

// CR 613.1f: a permanent that has lost all its abilities has no state
// triggers.
func TestStateTriggerNeedsTheAbility(t *testing.T) {
	withStateCatalog(t, stThresholdRow())
	g := newActiveGame(t)
	probe := pushStateProbe(t, g, "Enchantment")
	registerScopedEffectForTest(t, g, probe, []Mod{LoseAllAbilitiesMod()}, IndefiniteDuration())
	addCharge(t, g, probe, 3)
	g.WithWriteLock(func() { g.runStateChecksLocked() })
	if q, s := stItems(g, probe, stThresholdKey); q+s != 0 {
		t.Fatalf("a permanent with no abilities triggered (%d queued, %d stacked)", q, s)
	}
}

// CR 704.3: a state-based-action sweep is one event. A creature that
// dies in the same sweep as the last other creature its controller had
// never saw the state "you control no other creatures".
func TestStateTriggerDoesNotSeeHalfASimultaneousSweep(t *testing.T) {
	withStateCatalog(t, stCrocRow())
	for _, crocFirst := range []bool{true, false} {
		g := newActiveGame(t)
		me := g.Seats[0]
		other := uuid.New()
		croc := uuid.New()
		// Pushed without an event, so nothing asks the board until the
		// sweep: both orders of the battlefield are tried.
		pushCroc := func() {
			g.Battlefield.PushTop(Card{
				InstanceID: croc, Name: "State Croc", TypeLine: "Creature — Crocodile", OracleID: stOracle,
				Owner: me.ID, Controller: me.ID, Power: 5, Toughness: 5,
			})
		}
		if crocFirst {
			pushCroc()
			pushTestCreature(g, other, me, 2, 2)
		} else {
			pushTestCreature(g, other, me, 2, 2)
			pushCroc()
		}
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				c := &g.Battlefield.Cards[i]
				if c.InstanceID == croc || c.InstanceID == other {
					c.DamageMarked = 10
				}
			}
			g.runStateChecksLocked()
		})
		if q, s := stItems(g, croc, stCrocKey); q+s != 0 {
			t.Fatalf("crocFirst=%v: the crocodile triggered on a board half-way through one sweep (%d queued, %d stacked)", crocFirst, q, s)
		}
	}
}

// A queued state trigger is a keyed catalog item, so the table is a
// restore point, and the restored item still latches the ability.
func TestStateTriggerWaitingIsARestorePoint(t *testing.T) {
	withStateCatalog(t, stThresholdRow())
	g := newActiveGame(t)
	probe := pushStateProbe(t, g, "Enchantment")
	addCharge(t, g, probe, 2)
	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("a waiting state trigger blocks the restore point: %+v", snap.Continuations)
	}
	restored, err := throughJSON(t, snap).RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	restored.WithWriteLock(func() { restored.stateTriggersLocked() })
	if q, _ := stItems(restored, probe, stThresholdKey); q != 1 {
		t.Fatalf("queued after restore = %d, want 1: the restored item still latches the ability", q)
	}
	restored.WithWriteLock(func() {
		restored.runStateChecksLocked()
		restored.resolveTopAbilityLocked()
	})
	restored.ReadSnapshot(func() {
		if c := restored.findCardByIDLocked(probe); c == nil || c.Counters[CounterCharge] != 0 {
			t.Fatalf("the restored state trigger did not resolve: %+v", c)
		}
	})
}

// CR 510.2: combat damage is dealt simultaneously. The engine deals it
// attacker by attacker, and a defending player whose lifelinking blocker
// is processed before the unblocked attacker would pass through a life
// total the rules never see. A state trigger reading the life total must
// see only the settled one.
func TestStateTriggerDoesNotSeeHalfOfCombatDamage(t *testing.T) {
	g := newActiveGame(t)
	attacker, defender := g.Seats[0], g.Seats[1]
	threshold := defender.Life + 2
	const key = "State Probe — life threshold"
	withStateCatalog(t, TriggeredAbility{
		Key: key,
		State: func(g *Game, _ *Card, controller uuid.UUID) bool {
			p := g.playerByIDLocked(controller)
			return p != nil && p.Life >= threshold
		},
		Effect: func(*Game, *StackItem) error { return nil },
	})
	probe := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: probe, Name: "State Probe", TypeLine: "Enchantment", OracleID: stOracle,
		Owner: defender.ID, Controller: defender.ID,
	})
	blocked, unblocked, blocker := uuid.New(), uuid.New(), uuid.New()
	// The blocked attacker first, so its lifelinking blocker's damage is
	// dealt before the unblocked attacker's.
	pushTestCreature(g, blocked, attacker, 1, 10)
	pushTestCreature(g, unblocked, attacker, 2, 2)
	pushTestCreature(g, blocker, defender, 3, 10, "lifelink")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			switch c.InstanceID {
			case blocked, unblocked:
				c.AttackingTarget = defender.ID
			case blocker:
				c.BlockingTarget = blocked
			}
		}
		if g.blockedAttackers == nil {
			g.blockedAttackers = map[uuid.UUID]bool{}
		}
		g.blockedAttackers[blocked] = true
		g.assignAndDealCombatDamageLocked(CombatStepRegular)
	})
	if want := threshold - 1; defender.Life != want {
		t.Fatalf("defender life = %d, want %d (3 gained, 2 lost)", defender.Life, want)
	}
	if q, s := stItems(g, probe, key); q+s != 0 {
		t.Fatalf("the state trigger saw a life total half-way through combat damage (%d queued, %d stacked)", q, s)
	}
}
