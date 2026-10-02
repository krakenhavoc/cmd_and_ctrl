package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// becomes_target_batch_test.go — #1539, CR 603.3 / 603.3b / 603.3d. A
// triggered ability's targets are chosen as it is put on the stack
// (CR 603.3d). A "becomes the target" trigger that choice sets off
// triggered DURING the CR 603.3b placement, so it waits until the whole
// batch is on the stack and goes on above it ("then abilities that
// triggered during this process go on the stack"). It is never part of
// the batch that targeted: not in that batch's ordering prompt, and not
// placed under it by APNAP order.

const (
	btTargetedOracle = "test-1539-targets-a-creature"
	btWatcherOracle  = "test-1539-becomes-target-watcher"
)

// btCatalog extends the #1529 batch catalog with a cast trigger that
// targets a creature, and a watcher that triggers whenever a creature
// its controller controls becomes the target of anything — ward's and
// Monk Gyatso's event, with neither one's extra conditions.
func btCatalog(t *testing.T) {
	t.Helper()
	withCatalogTriggers(t, func(id string) []TriggeredAbility {
		switch id {
		case batchPlainOracle:
			return []TriggeredAbility{batchAbility(id)}
		case btTargetedOracle:
			return []TriggeredAbility{{
				Watches:   []EventKind{EventCast},
				AppliesTo: func(Event, *Card, Characteristic, *Game) bool { return true },
				Targets: &TargetSpec{
					Mode: "creature", Label: "target creature", Zones: []ZoneKind{ZoneBattlefield},
					CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool { return c.IsCreature() },
					Min:    1, Max: 1,
				},
				Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
					return newTriggeredItemForTest(source, btTargetedOracle, func(*Game, *StackItem) error { return nil })
				},
			}}
		case btWatcherOracle:
			return []TriggeredAbility{{
				Watches: []EventKind{EventBecomesTarget},
				AppliesTo: func(ev Event, source *Card, _ Characteristic, g *Game) bool {
					if ev.CardID == uuid.Nil {
						return false
					}
					c := g.findCardByIDLocked(ev.CardID)
					return c != nil && c.IsCreature() && c.Controller == source.Controller
				},
				Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
					return newTriggeredItemForTest(source, btWatcherOracle, func(*Game, *StackItem) error { return nil })
				},
			}}
		}
		return nil
	})
}

func pickCard(t *testing.T, g *Game, chooser, card uuid.UUID) {
	t.Helper()
	c := openPickTarget(g, chooser)
	if c == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(c.ID, chooser, TargetRef{Kind: TargetCard, ID: card}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
}

// TestBecomesTargetTriggerIsNotOrderedWithTheBatchThatTargeted — the
// issue's case. One player's batch is a plain trigger and a trigger that
// targets their own creature; they also control the watcher. The target
// pick sets the watcher off, but the ordering prompt offers only the two
// triggers of the batch, and the watcher's trigger goes on above both.
func TestBecomesTargetTriggerIsNotOrderedWithTheBatchThatTargeted(t *testing.T) {
	btCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	plain := pushBatchSource(g, me, batchPlainOracle)
	targeter := pushBatchSource(g, me, btTargetedOracle)
	watcher := pushBatchSource(g, me, btWatcherOracle)
	bear := pushBear(g, me.ID)

	fireBatch(g, me)
	pickCard(t, g, me.ID, bear)

	ch := findTriggerOrderPrompt(g, me.ID)
	if ch == nil {
		t.Fatal("no CR 603.3b ordering prompt for the plain and targeted triggers")
	}
	if w := pendingFrom(g, watcher); w != nil {
		t.Fatal("the becomes-the-target trigger joined the batch whose target choice set it off")
	}
	if len(ch.TriggerOrderIDs) != 2 {
		t.Fatalf("the prompt orders %d items, want the 2 triggers of the batch", len(ch.TriggerOrderIDs))
	}
	if err := g.ResolveTriggerOrder(ch.ID, me.ID, []uuid.UUID{pendingFrom(g, plain).ID, pendingFrom(g, targeter).ID}); err != nil {
		t.Fatalf("ResolveTriggerOrder: %v", err)
	}
	p, tg, w := stackItemFrom(g, plain), stackItemFrom(g, targeter), stackItemFrom(g, watcher)
	if p == nil || tg == nil || w == nil {
		t.Fatalf("after ordering: plain %v, targeted %v, watcher %v on the stack; want all three", p != nil, tg != nil, w != nil)
	}
	if w.Seq < p.Seq || w.Seq < tg.Seq {
		t.Errorf("the becomes-the-target trigger (seq %d) is not above the batch (plain %d, targeted %d)", w.Seq, p.Seq, tg.Seq)
	}
	if len(g.PendingTriggers) != 0 || findTriggerOrderPrompt(g, me.ID) != nil {
		t.Errorf("after placement: %d still queued, ordering prompt open %v", len(g.PendingTriggers), findTriggerOrderPrompt(g, me.ID) != nil)
	}
}

// TestBecomesTargetTriggerGoesAboveAnotherSeatsTargetingTrigger — the
// APNAP half, and ward's case. The non-active player's trigger targets
// the active player's creature. Had the watcher joined that batch, APNAP
// would have put the active player's trigger on the stack FIRST, under
// the trigger that targeted, and a ward counter would come too late.
func TestBecomesTargetTriggerGoesAboveAnotherSeatsTargetingTrigger(t *testing.T) {
	btCatalog(t)
	g := newActiveGame(t)
	ap, nap := g.Seats[g.Turn.ActiveSeat], g.Seats[1-g.Turn.ActiveSeat]
	targeter := pushBatchSource(g, nap, btTargetedOracle)
	watcher := pushBatchSource(g, ap, btWatcherOracle)
	bear := pushBear(g, ap.ID)

	fireBatch(g, ap)
	pickCard(t, g, nap.ID, bear)

	tg, w := stackItemFrom(g, targeter), stackItemFrom(g, watcher)
	if tg == nil || w == nil {
		t.Fatalf("after the pick: targeting trigger %v, watcher %v on the stack; want both", tg != nil, w != nil)
	}
	if w.Seq < tg.Seq {
		t.Errorf("the active player's becomes-the-target trigger (seq %d) is under the trigger that targeted (seq %d)", w.Seq, tg.Seq)
	}
}

// TestHeldTargetAnnouncementSurvivesUndoAndSnapshot — while the batch
// waits on its ordering prompt, the targeted item has chosen its target
// but has not yet been put on the stack, so its "becomes the target"
// event is still owed. An undo and a persisted restore point taken
// there must both still owe it.
func TestHeldTargetAnnouncementSurvivesUndoAndSnapshot(t *testing.T) {
	btCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushBatchSource(g, me, batchPlainOracle)
	targeter := pushBatchSource(g, me, btTargetedOracle)
	watcher := pushBatchSource(g, me, btWatcherOracle)
	bear := pushBear(g, me.ID)

	fireBatch(g, me)
	pickCard(t, g, me.ID, bear)
	if findTriggerOrderPrompt(g, me.ID) == nil {
		t.Fatal("no ordering prompt")
	}
	if it := pendingFrom(g, targeter); it == nil || !it.TargetsAnnouncePending {
		t.Fatal("the held targeted item does not owe its CR 603.3d announcement")
	}

	answer := func(t *testing.T, g *Game) {
		t.Helper()
		ch := findTriggerOrderPrompt(g, me.ID)
		if ch == nil {
			t.Fatal("no ordering prompt to answer")
		}
		if err := g.ResolveTriggerOrder(ch.ID, me.ID, ch.TriggerOrderIDs); err != nil {
			t.Fatalf("ResolveTriggerOrder: %v", err)
		}
		tg, w := stackItemFrom(g, targeter), stackItemFrom(g, watcher)
		if tg == nil || w == nil {
			t.Fatalf("targeted %v, watcher %v on the stack; want both", tg != nil, w != nil)
		}
		if tg.TargetsAnnouncePending {
			t.Error("the placed item still owes its announcement")
		}
		if w.Seq < tg.Seq {
			t.Errorf("watcher (seq %d) under the targeting trigger (seq %d)", w.Seq, tg.Seq)
		}
	}

	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	snap := g.Clone()

	t.Run("undo", func(t *testing.T) {
		answer(t, g)
		g.RestoreFrom(snap)
		if it := pendingFrom(g, targeter); it == nil || !it.TargetsAnnouncePending {
			t.Fatal("undo lost the owed announcement")
		}
		answer(t, g)
	})
	t.Run("snapshot", func(t *testing.T) {
		var s GameSnapshot
		if err := json.Unmarshal(raw, &s); err != nil {
			t.Fatal(err)
		}
		restored, err := s.Restore()
		if err != nil {
			t.Fatal(err)
		}
		if it := pendingFrom(restored, targeter); it == nil || !it.TargetsAnnouncePending {
			t.Fatal("the restore point lost the owed announcement")
		}
		answer(t, restored)
	})
}
