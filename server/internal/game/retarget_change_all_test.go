package game

import (
	"testing"
)

// retarget_change_all_test.go — #6228: CR 115.7a's "change the target or
// targets" (RetargetChangeAll) and the EventTargetsChosen it, and every
// first announcement, emits.

const targetsChosenWatcher = "test-6228-targets-chosen-watcher"

func targetsChosenEvents(g *Game) []Event {
	var out []Event
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == EventTargetsChosen {
				out = append(out, ev)
			}
		}
	})
	return out
}

func withTargetsChosenWatcher(t *testing.T, g *Game) {
	t.Helper()
	withCatalogTriggerRows(t, targetsChosenWatcher, []TriggeredAbility{{
		Key: "watch", Watches: []EventKind{EventTargetsChosen},
		AppliesTo: func(Event, *Card, Characteristic, *Game) bool { return false },
	}})
	pushTypedTestCard(g, Card{Name: "Watcher", TypeLine: "Enchantment", OracleID: targetsChosenWatcher,
		Owner: g.Seats[0].ID, Controller: g.Seats[0].ID})
}

// CR 115.7a: every target moves, each to ANOTHER legal target, staged
// and committed once.
func TestChangeAllMovesEveryTarget(t *testing.T) {
	g := newActiveGame(t)
	retargetSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushRetargetCreature(g, opp, "A")
	b := pushRetargetCreature(g, opp, "B")
	c := pushRetargetCreature(g, opp, "C")
	d := pushRetargetCreature(g, opp, "D")
	item := castTargeted(t, g, me, retargetPairOracle, []TargetRef{
		{Kind: TargetCard, ID: a, Slot: 0},
		{Kind: TargetCard, ID: b, Slot: 1},
	})
	g.WithWriteLock(func() {
		if err := g.OfferRetargetForEffect(RetargetOffer{ItemID: item.ID, Chooser: opp.ID, Policy: RetargetChangeAll, Optional: true}); err != nil {
			t.Fatalf("OfferRetargetForEffect: %v", err)
		}
	})
	first := findRetargetPrompt(g, opp.ID)
	if first == nil || first.RetargetSlot != 0 || first.PickTargetMin != 0 {
		t.Fatalf("first prompt = %+v, want an optional slot 0", first)
	}
	if hasUUID(first.PickTargetCards, a) {
		t.Error("slot 0 offers the target already in it (CR 115.7a: ANOTHER legal target)")
	}
	if err := g.ResolveRetarget(first.ID, opp.ID, []TargetRef{{Kind: TargetCard, ID: c}}); err != nil {
		t.Fatalf("slot 0: %v", err)
	}
	if got := g.StackMeta[item.ID].Targets; got[0].ID != a {
		t.Fatalf("slot 0 committed before the last slot was answered: %+v", got)
	}
	second := findRetargetPrompt(g, opp.ID)
	if second == nil || second.RetargetSlot != 1 {
		t.Fatalf("second prompt = %+v, want slot 1", second)
	}
	if err := g.ResolveRetarget(second.ID, opp.ID, []TargetRef{{Kind: TargetCard, ID: b}}); err != ErrIllegalTarget {
		t.Errorf("keeping slot 1's target: err = %v, want ErrIllegalTarget", err)
	}
	if err := g.ResolveRetarget(second.ID, opp.ID, []TargetRef{{Kind: TargetCard, ID: d}}); err != nil {
		t.Fatalf("slot 1: %v", err)
	}
	if got := g.StackMeta[item.ID].Targets; got[0].ID != c || got[1].ID != d {
		t.Errorf("targets = %+v, want [C D]", got)
	}
}

// CR 115.7a: "If all the targets aren't changed to other legal targets,
// none of them are changed." A declined second slot undoes the first.
func TestChangeAllDeclinedSlotChangesNothing(t *testing.T) {
	g := newActiveGame(t)
	retargetSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushRetargetCreature(g, opp, "A")
	b := pushRetargetCreature(g, opp, "B")
	c := pushRetargetCreature(g, opp, "C")
	pushRetargetCreature(g, opp, "D")
	item := castTargeted(t, g, me, retargetPairOracle, []TargetRef{
		{Kind: TargetCard, ID: a, Slot: 0},
		{Kind: TargetCard, ID: b, Slot: 1},
	})
	g.WithWriteLock(func() {
		_ = g.OfferRetargetForEffect(RetargetOffer{ItemID: item.ID, Chooser: opp.ID, Policy: RetargetChangeAll, Optional: true})
	})
	first := findRetargetPrompt(g, opp.ID)
	if err := g.ResolveRetarget(first.ID, opp.ID, []TargetRef{{Kind: TargetCard, ID: c}}); err != nil {
		t.Fatalf("slot 0: %v", err)
	}
	second := findRetargetPrompt(g, opp.ID)
	if err := g.ResolveRetarget(second.ID, opp.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if findRetargetPrompt(g, opp.ID) != nil {
		t.Error("a declined slot should end the walk")
	}
	if got := g.StackMeta[item.ID].Targets; got[0].ID != a || got[1].ID != b {
		t.Errorf("targets = %+v, want [A B] unchanged", got)
	}
}

// CR 115.7a: a slot with no other legal target means none changes, and
// nobody is asked.
func TestChangeAllWithAStuckSlotAsksNothing(t *testing.T) {
	g := newActiveGame(t)
	retargetSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushRetargetCreature(g, opp, "A")
	b := pushRetargetCreature(g, opp, "B")
	item := castTargeted(t, g, me, retargetPairOracle, []TargetRef{
		{Kind: TargetCard, ID: a, Slot: 0},
		{Kind: TargetCard, ID: b, Slot: 1},
	})
	g.WithWriteLock(func() {
		if err := g.OfferRetargetForEffect(RetargetOffer{ItemID: item.ID, Chooser: opp.ID, Policy: RetargetChangeAll, Optional: true}); err != nil {
			t.Fatalf("OfferRetargetForEffect: %v", err)
		}
	})
	// A and B can only swap, and a Distinct pair can't hold the same
	// creature twice, so slot 0 has nowhere to go that slot 1 allows.
	if p := findRetargetPrompt(g, opp.ID); p != nil {
		t.Errorf("asked %+v with no complete change possible", p)
	}
}

// The entry point refuses a ChangeAll list that leaves a slot alone.
func TestChangeAllRefusesAnUnchangedSlot(t *testing.T) {
	g := newActiveGame(t)
	retargetSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushRetargetCreature(g, opp, "A")
	b := pushRetargetCreature(g, opp, "B")
	c := pushRetargetCreature(g, opp, "C")
	item := castTargeted(t, g, me, retargetPairOracle, []TargetRef{
		{Kind: TargetCard, ID: a, Slot: 0},
		{Kind: TargetCard, ID: b, Slot: 1},
	})
	g.WithWriteLock(func() {
		err := g.RetargetStackItemForEffect(item.ID, opp.ID, RetargetChangeAll, []TargetRef{
			{Kind: TargetCard, ID: c, Slot: 0}, {Kind: TargetCard, ID: b, Slot: 1},
		})
		if err != ErrInvalidParam {
			t.Errorf("err = %v, want ErrInvalidParam", err)
		}
	})
}

// EventTargetsChosen: none without a watcher; one per announcement with
// one, for the cast and for a change of targets, the latter stamped with
// the changing card's name.
func TestTargetsChosenEventIsGatedAndStamped(t *testing.T) {
	g := newActiveGame(t)
	retargetSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushRetargetCreature(g, opp, "A")
	b := pushRetargetCreature(g, opp, "B")
	castTargeted(t, g, me, retargetOneOracle, []TargetRef{{Kind: TargetCard, ID: a}})
	if n := len(targetsChosenEvents(g)); n != 0 {
		t.Fatalf("%d targets-chosen events with no watcher, want 0", n)
	}

	withTargetsChosenWatcher(t, g)
	item := castTargeted(t, g, me, retargetPairOracle, []TargetRef{
		{Kind: TargetCard, ID: a, Slot: 0},
		{Kind: TargetCard, ID: b, Slot: 1},
	})
	evs := targetsChosenEvents(g)
	if len(evs) != 1 {
		t.Fatalf("%d targets-chosen events for one two-target cast, want 1", len(evs))
	}
	if evs[0].StackItemID != item.ID || evs[0].Actor != me.ID || evs[0].RetargetSourceName != "" {
		t.Errorf("cast event = %+v", evs[0])
	}

	c := pushRetargetCreature(g, opp, "C")
	src := pushTypedTestCard(g, Card{Name: "Redirector", TypeLine: "Enchantment", Owner: opp.ID, Controller: opp.ID})
	g.WithWriteLock(func() {
		if err := g.OfferRetargetForEffect(RetargetOffer{ItemID: item.ID, Chooser: opp.ID, Policy: RetargetChooseNew, Optional: true, Source: src}); err != nil {
			t.Fatalf("OfferRetargetForEffect: %v", err)
		}
	})
	p := findRetargetPrompt(g, opp.ID)
	if err := g.ResolveRetarget(p.ID, opp.ID, []TargetRef{{Kind: TargetCard, ID: c}}); err != nil {
		t.Fatalf("ResolveRetarget: %v", err)
	}
	evs = targetsChosenEvents(g)
	if len(evs) != 2 {
		t.Fatalf("%d targets-chosen events after the change, want 2", len(evs))
	}
	if evs[1].Actor != opp.ID || evs[1].RetargetSourceName != "Redirector" {
		t.Errorf("retarget event = %+v, want the chooser and the changing card's name", evs[1])
	}
}
