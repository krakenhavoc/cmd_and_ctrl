package game

import "testing"

// trigger_snapshot_x_test.go — #3782, CR 202.3e: "that spell's mana
// value" for a spell cast with X is the X on the stack, and the trigger
// snapshot is the one place it can be written down, because the spell is
// on the stack only until it resolves.

func TestACastTriggersSnapshotCountsTheXTheSpellWasCastWith(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	triggerCatalog(t, TriggeredAbility{
		Watches: []EventKind{EventCast},
		AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
			return ev.Actor == source.Controller
		},
		Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			return NewTriggeredItem(source, "watcher")
		},
	})
	watcherNamed(g, me.ID, "Watcher")
	spell := xSorcery(me.ID, "test-stroke") // {X}{G}
	me.Hand.PushTop(spell)

	if err := g.CastSpell(me.ID, spell.InstanceID, CastSpellParams{XValue: 3}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	item := onlyAbilityOnStack(g)
	if item == nil || item.Trigger == nil || item.Trigger.Object == nil {
		t.Fatalf("no cast trigger with a snapshot: %+v", item)
	}
	if got := item.Trigger.Object.ManaValue; got != 4 {
		t.Errorf("snapshot mana value = %d, want 4 for {X}{G} cast with X=3", got)
	}
}

// A permanent's {X} is zero everywhere but the stack, so the snapshot of
// one is unchanged by the X-aware read.
func TestASnapshotOfAPermanentReadsXAsZero(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	c := NewCard("Hangarback", me.ID)
	c.TypeLine = "Artifact Creature — Construct"
	c.ManaCost = "{X}{X}"
	c.Controller = me.ID
	id := pushPermanent(g, c)
	var mv int
	g.WithWriteLock(func() { mv = g.objectSnapshotLocked(id).ManaValue })
	if mv != 0 {
		t.Errorf("battlefield snapshot mana value = %d, want 0", mv)
	}
}
