package game

import (
	"testing"

	"github.com/google/uuid"
)

// cleanup_priority_test.go — CR 514.3a (#661). The cleanup step gives
// nobody priority (CR 514.3) right up until something happens in it:
// a state-based action performed, or a triggered ability waiting to
// go on the stack. Then the SBAs happen, the triggers go on the
// stack, the active player gets priority IN the cleanup step, and
// when the stack is empty and everyone has passed, another cleanup
// step begins.
//
// Before the fix the cleanup hook advanced out of the turn as soon as
// nobody owed a discard, so a trigger from the hand-size discard sat
// on PendingTriggers until the next player's upkeep.

// cleanupStepsBegun counts the EventStepBegan announcements for the
// cleanup step — the number of cleanup steps this game has had. The
// second one is CR 514.3a's whole observable payload.
func cleanupStepsBegun(g *Game) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == EventStepBegan && ev.Step == StepCleanup {
			n++
		}
	}
	return n
}

// closeCleanupWindow passes priority round the table once with an
// empty stack — "all players pass in succession", the pass that ends
// a CR 514.3a window and begins another cleanup step. Returns once
// the table has been round, which is when the engine has acted on it.
func closeCleanupWindow(t *testing.T, g *Game) {
	t.Helper()
	step, seat := g.Turn.Step, g.Turn.ActiveSeat
	for i := 0; i < 8; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
		if g.Turn.Step != step || g.Turn.ActiveSeat != seat {
			return
		}
		if g.Turn.PriorityHolder == g.Turn.ActiveSeat || g.Turn.PriorityHolder == NoPriority {
			return
		}
	}
	t.Fatalf("priority never came back round at %s", step)
}

// fillHandTo draws until the player holds n cards.
func fillHandTo(t *testing.T, g *Game, p *Player, n int) {
	t.Helper()
	for i := 0; p.Hand.Size() < n; i++ {
		if i > 40 {
			t.Fatalf("hand stuck at %d, want %d", p.Hand.Size(), n)
		}
		if err := g.DrawCard(p.ID); err != nil {
			t.Fatalf("DrawCard: %v", err)
		}
	}
}

// advanceIntoCleanup walks the cursor to the end step and takes one
// more step, which is the cleanup step's entry.
func advanceIntoCleanup(t *testing.T, g *Game) {
	t.Helper()
	advanceTo(t, g, StepEnd)
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep past End: %v", err)
	}
}

// TestCleanupWithNothingWaitingEndsTheTurn is the no-regression half
// of CR 514.3a: an uneventful cleanup grants nobody priority and the
// turn ends in the same call, exactly as it did before #661. The
// autopass and bot suites depend on no new stop appearing here.
func TestCleanupWithNothingWaitingEndsTheTurn(t *testing.T) {
	g := newActiveGame(t)
	advanceIntoCleanup(t, g)

	if g.Turn.ActiveSeat != 1 {
		t.Fatalf("turn did not end: active seat %d, want 1 (step %s)", g.Turn.ActiveSeat, g.Turn.Step)
	}
	if g.Turn.Step == StepCleanup {
		t.Errorf("cursor parked in cleanup with nothing waiting")
	}
	if n := cleanupStepsBegun(g); n != 1 {
		t.Errorf("cleanup steps begun = %d, want 1", n)
	}
}

// TestCleanupDiscardTriggerGetsPriorityAndASecondCleanupStep is the
// issue's repro in miniature: a permanent watching EventDiscardCard
// sees the CR 514.1 hand-size discard, and its trigger has to resolve
// in the turn that discarded — not in the next player's upkeep.
func TestCleanupDiscardTriggerGetsPriorityAndASecondCleanupStep(t *testing.T) {
	g := newActiveGame(t)
	active := g.Seats[0]
	const oracle = "test-cleanup-discard-watcher"
	watcher := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: watcher,
		Name:       "Test Discard Watcher",
		OracleID:   oracle,
		TypeLine:   "Creature — Test",
		Power:      3,
		Toughness:  3,
		Owner:      active.ID,
		Controller: active.ID,
	})
	withCatalogTriggers(t, func(id string) []TriggeredAbility {
		if id != oracle {
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventDiscardCard},
			AppliesTo: func(_ Event, _ *Card, _ Characteristic, _ *Game) bool {
				return true
			},
			Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				return NewTriggeredItem(source, "Test Discard Watcher — gain 3 life",
					func(g *Game, item *StackItem) error {
						return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, 3)
					})
			},
		}}
	})

	fillHandTo(t, g, active, 9)
	lifeBefore := active.Life
	advanceIntoCleanup(t, g)
	if g.Turn.Step != StepCleanup || g.DiscardPending[active.ID] != 2 {
		t.Fatalf("expected a cleanup discard pause, at %s pending=%v", g.Turn.Step, g.DiscardPending)
	}

	picks := []uuid.UUID{active.Hand.Cards[0].InstanceID, active.Hand.Cards[1].InstanceID}
	if err := g.DiscardSelection(active.ID, picks); err != nil {
		t.Fatalf("DiscardSelection: %v", err)
	}

	// CR 514.3a: still this turn's cleanup step, with the active
	// player on priority and the triggers on the stack.
	if g.Turn.Step != StepCleanup || g.Turn.ActiveSeat != 0 {
		t.Fatalf("left the turn after the discard: seat %d step %s", g.Turn.ActiveSeat, g.Turn.Step)
	}
	if g.Turn.PriorityHolder != 0 {
		t.Fatalf("PriorityHolder = %d in cleanup, want the active seat 0", g.Turn.PriorityHolder)
	}
	if n := len(g.StackMeta); n != 2 {
		t.Fatalf("stack items in cleanup = %d, want 2 (one per discarded card)", n)
	}
	if active.Life != lifeBefore {
		t.Fatalf("trigger resolved before anyone got priority: life %d, want %d", active.Life, lifeBefore)
	}

	// The table resolves them, still in cleanup, still this turn.
	settleStack(t, g)
	if active.Life != lifeBefore+6 {
		t.Errorf("life after both triggers: got %d, want %d", active.Life, lifeBefore+6)
	}
	if g.Turn.ActiveSeat != 0 || g.Turn.Step != StepCleanup {
		t.Fatalf("the turn moved on mid-resolution: seat %d step %s", g.Turn.ActiveSeat, g.Turn.Step)
	}

	// Stack empty, everyone passes: another cleanup step begins, has
	// nothing to do, and the turn ends.
	closeCleanupWindow(t, g)
	if n := cleanupStepsBegun(g); n != 2 {
		t.Errorf("cleanup steps begun = %d, want 2 (CR 514.3a's second one)", n)
	}
	if g.Turn.ActiveSeat != 1 {
		t.Errorf("turn did not end after the second cleanup: seat %d step %s", g.Turn.ActiveSeat, g.Turn.Step)
	}
}

// TestTriggerInTheSecondCleanupGivesAThird — CR 514.3a is not a
// one-shot. A trigger waiting in the repeated cleanup step earns
// another priority window and another cleanup step, for as long as
// the triggers keep coming.
func TestTriggerInTheSecondCleanupGivesAThird(t *testing.T) {
	g := newActiveGame(t)
	active := g.Seats[0]
	const oracle = "test-cleanup-step-watcher"
	watcher := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: watcher,
		Name:       "Test Cleanup Watcher",
		OracleID:   oracle,
		TypeLine:   "Enchantment",
		Owner:      active.ID,
		Controller: active.ID,
	})
	// "At the beginning of the cleanup step" — the trigger CR 514.3a
	// names explicitly. Bounded at two firings so the test ends: a
	// card that really did this every cleanup would be a CR 726 loop,
	// which is ADR 0055's problem and not this one's.
	fired := 0
	withCatalogTriggers(t, func(id string) []TriggeredAbility {
		if id != oracle {
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventStepBegan},
			AppliesTo: func(ev Event, _ *Card, _ Characteristic, _ *Game) bool {
				return ev.Step == StepCleanup && fired < 2
			},
			Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				fired++
				return NewTriggeredItem(source, "Test Cleanup Watcher — gain 1 life",
					func(g *Game, item *StackItem) error {
						return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, 1)
					})
			},
		}}
	})

	lifeBefore := active.Life
	advanceIntoCleanup(t, g)
	for i := 0; i < 3 && g.Turn.Step == StepCleanup; i++ {
		if g.Turn.PriorityHolder != g.Turn.ActiveSeat {
			t.Fatalf("cleanup %d: PriorityHolder = %d, want %d", i+1, g.Turn.PriorityHolder, g.Turn.ActiveSeat)
		}
		settleStack(t, g)
		closeCleanupWindow(t, g)
	}

	if fired != 2 {
		t.Fatalf("trigger fired %d times, want 2", fired)
	}
	if n := cleanupStepsBegun(g); n != 3 {
		t.Errorf("cleanup steps begun = %d, want 3 (two triggered, the third quiet)", n)
	}
	if active.Life != lifeBefore+2 {
		t.Errorf("life = %d, want %d — both cleanup triggers resolve in their own turn", active.Life, lifeBefore+2)
	}
	if g.Turn.ActiveSeat != 1 {
		t.Errorf("turn did not end after the quiet third cleanup: seat %d step %s", g.Turn.ActiveSeat, g.Turn.Step)
	}
}

// TestStateBasedActionInCleanupGrantsPriority is the other half of
// the CR 514.3a condition: no trigger at all, just a state-based
// action performed as a result of the check. The CR 514.2 sweep in
// this very step is what causes it — a creature kept alive by an
// "until end of turn" pump has 0 toughness the moment the pump ends.
//
// (The rule's other classic, a creature with lethal damage, cannot
// happen here: CR 514.2 removes marked damage before the check.)
func TestStateBasedActionInCleanupGrantsPriority(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	bear := pushScopedTestCreature(g, owner.ID, 2, 2)
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.New(), g.PinnedObjectsLocked(bear),
			[]Mod{ModifyPTMod(3, 3)}, g.UntilEndOfTurnDuration(), "test — +3/+3 until end of turn")
	})
	if err := g.AddCounter(bear, "-1/-1", 4); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	// 2/2, +3/+3 until end of turn, four -1/-1 counters: alive at 1/1
	// while the pump lasts, 0 or less the moment it ends. (Counters
	// live outside Effective(), so the toughness the SBA reads is
	// CurrentToughness, not the layer output.)
	var tough int
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == bear {
				tough = c.CurrentToughness()
			}
		}
	})
	if tough != 1 {
		t.Fatalf("setup toughness = %d, want 1 (2/2, +3/+3, four -1/-1)", tough)
	}

	advanceIntoCleanup(t, g)

	if g.Turn.Step != StepCleanup || g.Turn.ActiveSeat != 0 {
		t.Fatalf("no priority window for the SBA: seat %d step %s", g.Turn.ActiveSeat, g.Turn.Step)
	}
	if g.Turn.PriorityHolder != 0 {
		t.Errorf("PriorityHolder = %d, want the active seat 0", g.Turn.PriorityHolder)
	}
	if g.Battlefield.Contains(bear) {
		t.Errorf("the 0-toughness creature survived the cleanup SBA check (CR 704.5f)")
	}
	if !owner.Graveyard.Contains(bear) {
		t.Errorf("the creature did not reach its owner's graveyard")
	}

	closeCleanupWindow(t, g)
	if n := cleanupStepsBegun(g); n != 2 {
		t.Errorf("cleanup steps begun = %d, want 2", n)
	}
	if g.Turn.ActiveSeat != 1 {
		t.Errorf("turn did not end after the second cleanup: seat %d step %s", g.Turn.ActiveSeat, g.Turn.Step)
	}
}

// TestUndoAcrossTheCleanupPriorityWindow — the window is ordinary
// game state (the cursor's PriorityHolder and the items on the
// stack), so an undo taken inside it rewinds and replays like any
// other. No new field to classify: Turn and StackMeta are already
// carried by Clone / RestoreFrom.
func TestUndoAcrossTheCleanupPriorityWindow(t *testing.T) {
	g := newActiveGame(t)
	active := g.Seats[0]
	const oracle = "test-cleanup-undo-watcher"
	watcher := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: watcher,
		Name:       "Test Undo Watcher",
		OracleID:   oracle,
		TypeLine:   "Creature — Test",
		Power:      1,
		Toughness:  1,
		Owner:      active.ID,
		Controller: active.ID,
	})
	withCatalogTriggers(t, func(id string) []TriggeredAbility {
		if id != oracle {
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventDiscardCard},
			AppliesTo: func(_ Event, _ *Card, _ Characteristic, _ *Game) bool {
				return true
			},
			Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				return NewTriggeredItem(source, "Test Undo Watcher — gain 2 life",
					func(g *Game, item *StackItem) error {
						return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, 2)
					})
			},
		}}
	})

	fillHandTo(t, g, active, 8)
	advanceIntoCleanup(t, g)
	if err := g.DiscardSelection(active.ID, []uuid.UUID{active.Hand.Cards[0].InstanceID}); err != nil {
		t.Fatalf("DiscardSelection: %v", err)
	}
	if g.Turn.Step != StepCleanup || g.Turn.PriorityHolder != 0 {
		t.Fatalf("no cleanup priority window to undo across: step %s holder %d", g.Turn.Step, g.Turn.PriorityHolder)
	}
	lifeInWindow := active.Life
	snap := g.Clone()

	// Play on past the snapshot: resolve the trigger and end the turn.
	settleStack(t, g)
	closeCleanupWindow(t, g)
	if g.Turn.ActiveSeat != 1 || active.Life != lifeInWindow+2 {
		t.Fatalf("pre-undo play: seat %d life %d", g.Turn.ActiveSeat, active.Life)
	}

	g.WithWriteLock(func() { g.RestoreFrom(snap) })

	restored := g.Seats[0]
	if g.Turn.Step != StepCleanup || g.Turn.ActiveSeat != 0 {
		t.Fatalf("undo did not land back in cleanup: seat %d step %s", g.Turn.ActiveSeat, g.Turn.Step)
	}
	if g.Turn.PriorityHolder != 0 {
		t.Errorf("undo lost the cleanup priority window: holder %d", g.Turn.PriorityHolder)
	}
	if restored.Life != lifeInWindow {
		t.Errorf("undo did not rewind life: got %d, want %d", restored.Life, lifeInWindow)
	}
	if len(g.StackMeta) != 1 {
		t.Fatalf("undo did not restore the waiting trigger: %d items", len(g.StackMeta))
	}
	// And it replays from there.
	settleStack(t, g)
	closeCleanupWindow(t, g)
	if restored.Life != lifeInWindow+2 {
		t.Errorf("replay after undo: life %d, want %d", restored.Life, lifeInWindow+2)
	}
	if g.Turn.ActiveSeat != 1 {
		t.Errorf("replay after undo did not end the turn: seat %d step %s", g.Turn.ActiveSeat, g.Turn.Step)
	}
}

// TestUntilEndOfTurnEffectMadeInCleanupExpiresInTheSecondCleanup is
// ADR 0035 §3's simplification, closed. An effect created during the
// cleanup step used to leak into the next turn: the CR 514.2 sweep
// for that turn had already run and there was no second cleanup step
// to run it again. There is now, and the expiry stamp does the rest —
// the effect is stamped with this turn's number, and the repeated
// cleanup drops everything stamped at or before it.
func TestUntilEndOfTurnEffectMadeInCleanupExpiresInTheSecondCleanup(t *testing.T) {
	g := newActiveGame(t)
	active := g.Seats[0]
	const oracle = "test-cleanup-pump-watcher"
	watcher := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: watcher,
		Name:       "Test Pump Watcher",
		OracleID:   oracle,
		TypeLine:   "Creature — Test",
		Power:      2,
		Toughness:  2,
		Owner:      active.ID,
		Controller: active.ID,
	})
	withCatalogTriggers(t, func(id string) []TriggeredAbility {
		if id != oracle {
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventDiscardCard},
			AppliesTo: func(_ Event, _ *Card, _ Characteristic, _ *Game) bool {
				return true
			},
			Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				return NewTriggeredItem(source, "Test Pump Watcher — +2/+2 until end of turn",
					func(g *Game, item *StackItem) error {
						self := item.SourceCardID
						g.RegisterScopedEffectForEffect(self, g.PinnedObjectsLocked(self),
							[]Mod{ModifyPTMod(2, 2)}, g.UntilEndOfTurnDuration(), "test — +2/+2 until end of turn")
						return nil
					})
			},
		}}
	})

	fillHandTo(t, g, active, 8)
	advanceIntoCleanup(t, g)
	if err := g.DiscardSelection(active.ID, []uuid.UUID{active.Hand.Cards[0].InstanceID}); err != nil {
		t.Fatalf("DiscardSelection: %v", err)
	}
	settleStack(t, g)
	if len(g.ScopedEffects) != 1 {
		t.Fatalf("the trigger did not create its effect in the cleanup step: %d registered", len(g.ScopedEffects))
	}

	closeCleanupWindow(t, g)

	if n := cleanupStepsBegun(g); n != 2 {
		t.Fatalf("cleanup steps begun = %d, want 2", n)
	}
	if len(g.ScopedEffects) != 0 {
		t.Errorf("the effect leaked into the next turn: %d still registered", len(g.ScopedEffects))
	}
	if g.Turn.ActiveSeat != 1 {
		t.Errorf("turn did not end: seat %d step %s", g.Turn.ActiveSeat, g.Turn.Step)
	}
}
