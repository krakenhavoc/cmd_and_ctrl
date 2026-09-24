package game

import (
	"testing"

	"github.com/google/uuid"
)

// source_new_object_test.go — #1432, CR 400.7. The engine half of the
// "this permanent after a flicker" guard: AbilitySourceIsNewObjectForEffect
// is the question every effects primitive asks about its own source,
// and followedSourceObjectLocked is what a delayed or reflexive trigger
// inherits when the resolution that creates it has moved its source
// onto the battlefield. The card-level proof is in
// cards/effects/source_object_primitives_test.go.

// resolveProbe pushes a stamped ability item from `src`, runs `before`
// (the response window), resolves it through the real resolution path,
// and returns what `during` saw. `during` runs inside the Effect, under
// the lock, with the resolving slot set exactly as a card's body sees
// it.
func resolveProbe(t *testing.T, g *Game, src uuid.UUID, before func(), during func(g *Game, item *StackItem)) {
	t.Helper()
	me := g.Seats[0]
	stamp := refOf(t, g, src)
	item := pushAbilityItem(g, StackItemTriggered, me.ID, src, "probe", func(g *Game, item *StackItem) error {
		during(g, item)
		return nil
	})
	g.WithWriteLock(func() { item.SourceObject = stamp })
	if before != nil {
		before()
	}
	g.WithWriteLock(func() { g.resolveTopAbilityLocked() })
}

// rawFlickerLocked is flickerRaw for a body that already holds the
// lock: the card leaves the battlefield and comes straight back.
func rawFlickerLocked(t *testing.T, g *Game, owner *Player, id uuid.UUID) {
	t.Helper()
	if _, err := MoveCard(g.Battlefield, owner.Hand, id); err != nil {
		t.Fatalf("bounce: %v", err)
	}
	if _, err := MoveCard(owner.Hand, g.Battlefield, id); err != nil {
		t.Fatalf("return: %v", err)
	}
}

// The control: a source that never moved is the object its ability
// names.
func TestALiveSourceIsNotANewObject(t *testing.T) {
	g := newActiveGame(t)
	src := watcherNamed(g, g.Seats[0].ID, "Relic")
	var got bool
	resolveProbe(t, g, src, nil, func(g *Game, item *StackItem) {
		got = g.AbilitySourceIsNewObjectForEffect(item)
	})
	if got {
		t.Error("a source that never moved reads as a new object")
	}
}

// The headline: flickered in response, the permanent under the same ID
// is a new object, and the resolving ability is told so.
func TestASourceFlickeredInResponseIsANewObject(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := watcherNamed(g, me.ID, "Relic")
	var got bool
	resolveProbe(t, g, src, func() { flickerRaw(t, g, me.ID, src) }, func(g *Game, item *StackItem) {
		got = g.AbilitySourceIsNewObjectForEffect(item)
	})
	if !got {
		t.Error("a source that left and came back in response reads as the object the ability names (CR 400.7)")
	}
}

// The exception: an effect that moves its own source can find the
// object it moved. A self-flicker that then acts on "it" acts.
func TestASourceThisResolutionMovedIsNotANewObject(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := watcherNamed(g, me.ID, "Relic")
	var got bool
	resolveProbe(t, g, src, nil, func(g *Game, item *StackItem) {
		rawFlickerLocked(t, g, me, src)
		got = g.AbilitySourceIsNewObjectForEffect(item)
	})
	if got {
		t.Error("the object this resolution itself put onto the battlefield reads as a stranger")
	}
}

// Off the battlefield the guard does not speak: "return this card from
// your graveyard" and a dies trigger's "return it" follow the card.
func TestASourceOffTheBattlefieldIsNotANewObject(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := watcherNamed(g, me.ID, "Relic")
	var got bool
	resolveProbe(t, g, src, func() {
		g.WithWriteLock(func() {
			if _, err := MoveCard(g.Battlefield, me.Graveyard, src); err != nil {
				t.Fatal(err)
			}
		})
	}, func(g *Game, item *StackItem) {
		got = g.AbilitySourceIsNewObjectForEffect(item)
	})
	if got {
		t.Error("a source in the graveyard reads as a new permanent")
	}
}

// Nothing to judge: a spell, an item with no stamp (a pre-#1418
// snapshot), and no item at all all keep today's behaviour.
func TestUnstampedItemsAndSpellsAreNeverANewObject(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := watcherNamed(g, me.ID, "Relic")
	first := refOf(t, g, src)
	flickerRaw(t, g, me.ID, src)
	g.WithWriteLock(func() {
		if g.AbilitySourceIsNewObjectForEffect(nil) {
			t.Error("nil item")
		}
		unstamped := &StackItem{ID: uuid.New(), Kind: StackItemTriggered, SourceCardID: src}
		if g.AbilitySourceIsNewObjectForEffect(unstamped) {
			t.Error("an unstamped item must keep the pre-#1418 behaviour")
		}
		spell := &StackItem{ID: uuid.New(), Kind: StackItemSpell, SourceCardID: src, SourceObject: first}
		if g.AbilitySourceIsNewObjectForEffect(spell) {
			t.Error("a spell's \"this\" is the spell, not a permanent")
		}
		// Outside any resolution the stamp alone decides.
		stamped := &StackItem{ID: uuid.New(), Kind: StackItemActivated, SourceCardID: src, SourceObject: first}
		if !g.AbilitySourceIsNewObjectForEffect(stamped) {
			t.Error("a stamped item whose source came back reads as the same object")
		}
	})
}

// Unearth's shape (CR 603.7d): an ability activated from the graveyard
// returns its card and schedules "exile it at the next end step". The
// delayed trigger is about the creature it returned, not the graveyard
// card the ability was stamped with — or it would never be exiled.
func TestADelayedTriggerFollowsTheSourceItsCreatorPutOntoTheBattlefield(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := watcherNamed(g, me.ID, "Relic")
	g.WithWriteLock(func() {
		if _, err := MoveCard(g.Battlefield, me.Graveyard, src); err != nil {
			t.Fatal(err)
		}
	})
	var graveRef ObjectRef
	g.WithWriteLock(func() { graveRef = g.sourceObjectRefLocked(src) })
	item := pushAbilityItem(g, StackItemActivated, me.ID, src, "unearth-ish", func(g *Game, item *StackItem) error {
		if _, err := MoveCard(me.Graveyard, g.Battlefield, src); err != nil {
			return err
		}
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller: me.ID, SourceCardID: src, Label: "exile it", At: StepEnd, Cards: []uuid.UUID{src},
			Body: testBody(func(*Game, *StackItem) error { return nil }),
		})
		if !queueReflexiveProbe(g, item) {
			t.Error("QueueReflexiveTriggerForEffect refused a well-formed declaration")
		}
		return nil
	})
	g.WithWriteLock(func() {
		item.SourceObject = graveRef
		g.resolveTopAbilityLocked()
	})
	returned := refOf(t, g, src)
	if returned == graveRef {
		t.Fatal("setup: the returned card should be a new object")
	}
	if len(g.DelayedTriggers) != 1 || g.DelayedTriggers[0].SourceObject != returned {
		t.Fatalf("delayed trigger = %+v, want SourceObject %+v (the returned permanent)", g.DelayedTriggers, returned)
	}
	if reflexive := queuedTrigger(g); reflexive == nil || reflexive.SourceObject != returned {
		t.Errorf("reflexive trigger = %+v, want SourceObject %+v", reflexive, returned)
	}

	// Fired, it knows its creature — and a creature flickered before the
	// end step is a new object the delayed trigger has no memory of.
	g.WithWriteLock(func() {
		g.PendingTriggers = nil
		g.fireDelayedTriggersLocked(StepEnd)
	})
	fired := queuedTrigger(g)
	if fired == nil {
		t.Fatal("the delayed trigger did not fire")
	}
	g.WithWriteLock(func() {
		if g.AbilitySourceIsNewObjectForEffect(fired) {
			t.Error("the fired delayed trigger does not recognise the creature its creator returned")
		}
	})
	flickerRaw(t, g, me.ID, src)
	g.WithWriteLock(func() {
		if !g.AbilitySourceIsNewObjectForEffect(fired) {
			t.Error("a creature flickered before the end step is still \"it\" to the delayed trigger")
		}
	})
}

// A move the resolution made to anywhere BUT the battlefield keeps the
// stamp, so "exile this. When you do, …" still reads the permanent it
// was through SourcePermanent's last-known record.
func TestADelayedTriggerKeepsTheStampWhenItsCreatorMovedTheSourceElsewhere(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := watcherNamed(g, me.ID, "Relic")
	stamp := refOf(t, g, src)
	item := pushAbilityItem(g, StackItemActivated, me.ID, src, "exile-ish", func(g *Game, item *StackItem) error {
		if _, err := MoveCard(g.Battlefield, me.Graveyard, src); err != nil {
			return err
		}
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller: me.ID, SourceCardID: src, Label: "later", At: StepEnd,
			Body: testBody(func(*Game, *StackItem) error { return nil }),
		})
		return nil
	})
	g.WithWriteLock(func() {
		item.SourceObject = stamp
		g.resolveTopAbilityLocked()
	})
	if len(g.DelayedTriggers) != 1 || g.DelayedTriggers[0].SourceObject != stamp {
		t.Fatalf("delayed trigger = %+v, want the stamp %+v", g.DelayedTriggers, stamp)
	}
}

// queueReflexiveProbe queues a no-op reflexive trigger from `parent`
// and reports whether the declaration was accepted.
func queueReflexiveProbe(g *Game, parent *StackItem) bool {
	return g.QueueReflexiveTriggerForEffect(parent, ReflexiveTrigger{
		Label: "when you do", Effect: func(*Game, *StackItem) error { return nil },
	})
}
