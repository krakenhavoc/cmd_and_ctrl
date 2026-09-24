package game

import (
	"testing"

	"github.com/google/uuid"
)

// trigger_batch_hold_test.go — #1529, CR 603.3b / 603.3d. A triggered
// ability that is still being put on the stack (its CR 603.5 "you may",
// its CR 603.3c mode pick or its CR 603.3d target pick is open) belongs
// to the batch it fired with. The drain holds the WHOLE trigger queue
// until it joins, so its controller orders it with the rest and the
// APNAP placement sees every seat's complete batch.

const (
	batchPlainOracle     = "test-1529-plain"
	batchTargetedOracle  = "test-1529-targeted"
	batchOptionalOracle  = "test-1529-optional"
	batchModalOracle     = "test-1529-modal"
	batchTwoClauseOracle = "test-1529-two-clause"
)

// batchAbility fires on every EventCast and builds an item labelled
// with the oracle ID, so a test can find which source's item is where.
func batchAbility(oracle string) TriggeredAbility {
	ab := TriggeredAbility{
		Watches:   []EventKind{EventCast},
		AppliesTo: func(Event, *Card, Characteristic, *Game) bool { return true },
		Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			return NewTriggeredItem(source, oracle, func(*Game, *StackItem) error { return nil })
		},
	}
	switch oracle {
	case batchTargetedOracle:
		ab.Targets = &TargetSpec{Mode: "player", Label: "target player", Players: true, Min: 1, Max: 1}
	case batchOptionalOracle:
		ab.OptionalPrompt = &TriggerOptionalPrompt{Question: "Fire the optional trigger?"}
	case batchTwoClauseOracle:
		// "Target creature and a second target creature." Each clause
		// is fillable on its own, so the harvest opens the walk; with
		// one creature on the board the second is not, once the first
		// has taken it.
		creature := func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool { return c.IsCreature() }
		ab.Targets = &TargetSpec{
			Mode: "creature", Label: "target creature", Zones: []ZoneKind{ZoneBattlefield},
			CardOK: creature, Min: 1, Max: 1,
			Rest: []TargetClause{{
				Mode: "creature", Label: "a second target creature", Zones: []ZoneKind{ZoneBattlefield},
				CardOK: creature, Min: 1, Max: 1, Distinct: true,
			}},
		}
	case batchModalOracle:
		noop := func(*Game, *StackItem, int) error { return nil }
		ab.Modes = &ModeSpec{Prompt: "Choose one", Min: 1, Max: 1, Options: []ModeOption{
			{Label: "One.", Effect: noop},
			{Label: "Two.", Effect: noop},
		}}
	}
	return ab
}

func withBatchCatalog(t *testing.T) {
	t.Helper()
	withCatalogTriggers(t, func(id string) []TriggeredAbility {
		switch id {
		case batchPlainOracle, batchTargetedOracle, batchOptionalOracle, batchModalOracle, batchTwoClauseOracle:
			return []TriggeredAbility{batchAbility(id)}
		}
		return nil
	})
}

func pushBatchSource(g *Game, owner *Player, oracle string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id, Name: oracle, OracleID: oracle, TypeLine: "Enchantment",
		Owner: owner.ID, Controller: owner.ID,
	})
	return id
}

// fireBatch emits one cast and runs the priority-grant boundary after
// it, the way a real cast does.
func fireBatch(g *Game, actor *Player) {
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventCast, Actor: actor.ID})
		g.runStateChecksLocked()
	})
}

func stackItemFrom(g *Game, source uuid.UUID) *StackItem {
	for _, it := range g.StackMeta {
		if it != nil && it.Kind == StackItemTriggered && it.SourceCardID == source {
			return it
		}
	}
	return nil
}

func pendingFrom(g *Game, source uuid.UUID) *StackItem {
	for _, it := range g.PendingTriggers {
		if it != nil && it.SourceCardID == source {
			return it
		}
	}
	return nil
}

func openTriggerPrompt(g *Game, chooser uuid.UUID) *PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == PendingChoiceTriggerPrompt && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

func triggeredOnStack(g *Game) int {
	n := 0
	for _, it := range g.StackMeta {
		if it != nil && it.Kind == StackItemTriggered {
			n++
		}
	}
	return n
}

// TestAnnouncingTriggerHoldsItsBatch — for each announcement prompt,
// the untargeted batch-mate waits in the queue while it is open, and
// once it is answered the two are ordered together.
func TestAnnouncingTriggerHoldsItsBatch(t *testing.T) {
	for _, tc := range []struct {
		oracle string
		answer func(t *testing.T, g *Game, me, opp *Player)
	}{
		{batchTargetedOracle, func(t *testing.T, g *Game, me, opp *Player) {
			c := openPickTarget(g, me.ID)
			if c == nil {
				t.Fatal("no target prompt")
			}
			if err := g.ResolvePickTarget(c.ID, me.ID, TargetRef{Kind: TargetPlayer, ID: opp.ID}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
		}},
		{batchOptionalOracle, func(t *testing.T, g *Game, me, _ *Player) {
			c := openTriggerPrompt(g, me.ID)
			if c == nil {
				t.Fatal("no optional-trigger prompt")
			}
			if err := g.ResolveTriggerPrompt(c.ID, me.ID, true); err != nil {
				t.Fatalf("ResolveTriggerPrompt: %v", err)
			}
		}},
		{batchModalOracle, func(t *testing.T, g *Game, me, _ *Player) {
			c := openModePick(t, g, me.ID)
			if c == nil {
				t.Fatal("no mode prompt")
			}
			if err := g.ResolveModePick(c.ID, me.ID, []int{1}); err != nil {
				t.Fatalf("ResolveModePick: %v", err)
			}
		}},
	} {
		t.Run(tc.oracle, func(t *testing.T) {
			withBatchCatalog(t)
			g := newActiveGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1-g.Turn.ActiveSeat]
			plain := pushBatchSource(g, me, batchPlainOracle)
			other := pushBatchSource(g, me, tc.oracle)

			fireBatch(g, me)
			if n := triggeredOnStack(g); n != 0 {
				t.Fatalf("%d trigger(s) reached the stack while a batch-mate was still announcing", n)
			}
			if pendingFrom(g, plain) == nil {
				t.Fatal("the plain trigger is not waiting in the queue")
			}
			if findTriggerOrderPrompt(g, me.ID) != nil {
				t.Fatal("the ordering prompt opened before the announcing trigger joined")
			}

			tc.answer(t, g, me, opp)

			ch := findTriggerOrderPrompt(g, me.ID)
			if ch == nil {
				t.Fatal("no CR 603.3b ordering prompt after the answer")
			}
			item := pendingFrom(g, other)
			if item == nil {
				t.Fatal("the announced trigger did not join the queue")
			}
			if len(ch.TriggerOrderIDs) != 2 {
				t.Fatalf("the prompt orders %d items, want 2", len(ch.TriggerOrderIDs))
			}
			// The announced trigger resolves LAST: under the plain one.
			if err := g.ResolveTriggerOrder(ch.ID, me.ID, []uuid.UUID{pendingFrom(g, plain).ID, item.ID}); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			p, o := stackItemFrom(g, plain), stackItemFrom(g, other)
			if p == nil || o == nil {
				t.Fatalf("after ordering: plain on stack %v, announced on stack %v", p != nil, o != nil)
			}
			if p.Seq < o.Seq {
				t.Errorf("the announced trigger is on top (seq %d over %d) though it was ordered to resolve last", o.Seq, p.Seq)
			}
		})
	}
}

// TestDecliningAnOptionalTriggerReleasesItsBatch — "no" still ends the
// announcement, and the held batch goes on the stack there and then.
func TestDecliningAnOptionalTriggerReleasesItsBatch(t *testing.T) {
	withBatchCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	plain := pushBatchSource(g, me, batchPlainOracle)
	pushBatchSource(g, me, batchOptionalOracle)

	fireBatch(g, me)
	c := openTriggerPrompt(g, me.ID)
	if c == nil || stackItemFrom(g, plain) != nil {
		t.Fatal("the plain trigger was not held behind the optional prompt")
	}
	if err := g.ResolveTriggerPrompt(c.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveTriggerPrompt: %v", err)
	}
	if stackItemFrom(g, plain) == nil || len(g.PendingTriggers) != 0 {
		t.Fatalf("after declining: plain on the stack %v, %d still queued", stackItemFrom(g, plain) != nil, len(g.PendingTriggers))
	}
	if findTriggerOrderPrompt(g, me.ID) != nil {
		t.Error("a one-item batch asked to be ordered")
	}
}

// TestTargetWalkRemovedMidWayReleasesItsBatch — CR 603.3d removes a
// trigger whose later required clause has nothing left once an earlier
// clause is answered. No item joins, so the answer itself has to run the
// boundary that releases the batch it was holding.
func TestTargetWalkRemovedMidWayReleasesItsBatch(t *testing.T) {
	withBatchCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	plain := pushBatchSource(g, me, batchPlainOracle)
	walker := pushBatchSource(g, me, batchTwoClauseOracle)
	bear := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
	})

	fireBatch(g, me)
	c := openPickTarget(g, me.ID)
	if c == nil || stackItemFrom(g, plain) != nil {
		t.Fatal("the plain trigger was not held behind the two-clause walk")
	}
	if err := g.ResolvePickTarget(c.ID, me.ID, TargetRef{Kind: TargetCard, ID: bear}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	if openPickTarget(g, me.ID) != nil {
		t.Fatal("the walk asked for a second creature that does not exist")
	}
	if triggerItemFrom(g, walker) != nil {
		t.Fatal("CR 603.3d: the trigger whose second clause emptied was put on the stack")
	}
	if stackItemFrom(g, plain) == nil || len(g.PendingTriggers) != 0 {
		t.Errorf("after the walk was removed: plain on the stack %v, %d still queued", stackItemFrom(g, plain) != nil, len(g.PendingTriggers))
	}
}

// TestAnnouncingTriggerHoldsEverySeatForAPNAP — the hold is the whole
// queue, not the announcing seat's: the active player's targeted trigger
// and the next player's plain one go on the stack together, active
// player's first, so the non-active player's resolves first (CR 603.3b).
// Before #1529 the plain one drained alone and the active player's
// landed above it.
func TestAnnouncingTriggerHoldsEverySeatForAPNAP(t *testing.T) {
	withBatchCatalog(t)
	g := newFourPlayerActiveGame(t)
	ap := g.Seats[g.Turn.ActiveSeat]
	nap := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	apSrc := pushBatchSource(g, ap, batchTargetedOracle)
	napSrc := pushBatchSource(g, nap, batchPlainOracle)

	fireBatch(g, ap)
	if stackItemFrom(g, napSrc) != nil {
		t.Fatal("the non-active player's trigger drained while the active player was choosing a target")
	}
	c := openPickTarget(g, ap.ID)
	if c == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(c.ID, ap.ID, TargetRef{Kind: TargetPlayer, ID: nap.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	a, n := stackItemFrom(g, apSrc), stackItemFrom(g, napSrc)
	if a == nil || n == nil {
		t.Fatalf("after the pick: AP item on stack %v, NAP item on stack %v", a != nil, n != nil)
	}
	if n.Seq < a.Seq {
		t.Errorf("APNAP: the active player's trigger (seq %d) is above the non-active player's (seq %d)", a.Seq, n.Seq)
	}
}

// TestConcedeReleasesABatchHeldForTheConcedersTrigger — a departure that
// drops the last announcement prompt the drain was holding for puts the
// rest of the batch on the stack at once, rather than leaving it queued
// until some unrelated action runs the boundary.
func TestConcedeReleasesABatchHeldForTheConcedersTrigger(t *testing.T) {
	withBatchCatalog(t)
	g := newFourPlayerActiveGame(t)
	ap := g.Seats[g.Turn.ActiveSeat]
	leaver := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	apSrc := pushBatchSource(g, ap, batchPlainOracle)
	pushBatchSource(g, leaver, batchTargetedOracle)

	fireBatch(g, ap)
	if openPickTarget(g, leaver.ID) == nil || stackItemFrom(g, apSrc) != nil {
		t.Fatal("the active player's trigger was not held behind the other seat's target prompt")
	}
	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if openPickTarget(g, leaver.ID) != nil {
		t.Fatal("the departed player's target prompt survived the concession")
	}
	if stackItemFrom(g, apSrc) == nil || len(g.PendingTriggers) != 0 {
		t.Errorf("after the concession: AP trigger on the stack %v, %d still queued", stackItemFrom(g, apSrc) != nil, len(g.PendingTriggers))
	}
}

// TestPickTargetFrameIsNotSharedWithAnUndoSnapshot — answering a target
// prompt advances its walk in place, so the undo snapshot needs its own
// frame or the restored prompt comes back already answered.
func TestPickTargetFrameIsNotSharedWithAnUndoSnapshot(t *testing.T) {
	withBatchCatalog(t)
	g := newActiveGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1-g.Turn.ActiveSeat]
	src := pushBatchSource(g, me, batchTargetedOracle)

	fireBatch(g, me)
	snap := g.Clone()
	c := openPickTarget(g, me.ID)
	if c == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(c.ID, me.ID, TargetRef{Kind: TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	if stackItemFrom(g, src) == nil {
		t.Fatal("the answered trigger did not reach the stack")
	}

	g.RestoreFrom(snap)
	c = openPickTarget(g, me.ID)
	if c == nil {
		t.Fatal("the restored game has no target prompt")
	}
	if f := c.pickTargetResume; f == nil || f.step != 0 || len(f.picked) != 0 {
		t.Fatalf("the restored walk is already answered: %+v", f)
	}
	if err := g.ResolvePickTarget(c.ID, me.ID, TargetRef{Kind: TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget after undo: %v", err)
	}
	it := stackItemFrom(g, src)
	if it == nil || len(it.Targets) != 1 || it.Targets[0].ID != opp.ID {
		t.Fatalf("after undo and a second answer: item %+v, want it on the stack targeting the opponent", it)
	}
}
