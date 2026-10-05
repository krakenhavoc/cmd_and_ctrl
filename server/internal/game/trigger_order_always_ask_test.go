package game

import (
	"encoding/json"
	"testing"
)

// trigger_order_always_ask_test.go — #1530: the per-seat "always ask me
// to order my triggers" preference restores the CR 603.3b prompt for a
// batch #1511 would auto-order.

func TestAlwaysAskDefaultOffKeepsAutoOrdering(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	if me.TriggerOrderAlwaysAsk {
		t.Fatal("the preference defaults to on")
	}
	queueProwessForTest(g, me)
	queueProwessForTest(g, me)
	g.WithWriteLock(func() {
		if !g.drainPendingTriggersAPNAPLocked() {
			t.Error("default-off seat was held behind an ordering prompt")
		}
	})
}

func TestAlwaysAskPromptsAnAllProwessBatch(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	if err := g.SetTriggerOrderPreference(me.ID, true); err != nil {
		t.Fatal(err)
	}
	queueProwessForTest(g, me)
	queueProwessForTest(g, me)
	queueProwessForTest(g, me)
	g.WithWriteLock(func() {
		if g.drainPendingTriggersAPNAPLocked() {
			t.Error("an always-ask seat's prowess batch drained without a prompt")
		}
	})
	prompt := findTriggerOrderPrompt(g, me.ID)
	if prompt == nil || len(prompt.TriggerOrderIDs) != 3 {
		t.Fatalf("want a 3-item trigger_order prompt, got %+v", prompt)
	}
}

// Only the seat that set it is affected.
func TestAlwaysAskIsPerSeat(t *testing.T) {
	g := newActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	if err := g.SetTriggerOrderPreference(a.ID, true); err != nil {
		t.Fatal(err)
	}
	queueProwessForTest(g, b)
	queueProwessForTest(g, b)
	g.WithWriteLock(func() {
		if !g.drainPendingTriggersAPNAPLocked() {
			t.Error("seat B was held by seat A's preference")
		}
	})
	if findTriggerOrderPrompt(g, b.ID) != nil {
		t.Error("seat B was prompted")
	}
}

// An answered batch (every item Ordered) is not asked twice.
func TestAlwaysAskDoesNotRepromptAnOrderedBatch(t *testing.T) {
	items := []*StackItem{{Commutes: true, Ordered: true}, {Commutes: true, Ordered: true}}
	if seatNeedsTriggerOrder(items, true) {
		t.Error("an Ordered batch re-prompted")
	}
	items[1].Ordered = false
	if !seatNeedsTriggerOrder(items, true) {
		t.Error("a batch with an unordered item did not prompt")
	}
}

func TestAlwaysAskRoundTripsCloneAndSnapshot(t *testing.T) {
	g := newActiveGame(t)
	if err := g.SetTriggerOrderPreference(g.Seats[1].ID, true); err != nil {
		t.Fatal(err)
	}
	if !g.Clone().Seats[1].TriggerOrderAlwaysAsk || g.Clone().Seats[0].TriggerOrderAlwaysAsk {
		t.Error("Clone lost or smeared the preference")
	}
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	r, err := snap.Restore()
	if err != nil {
		t.Fatal(err)
	}
	if !r.Seats[1].TriggerOrderAlwaysAsk || r.Seats[0].TriggerOrderAlwaysAsk {
		t.Error("snapshot round trip lost or smeared the preference")
	}
}

// Undo is Clone + RestoreFrom. The preference is a setting, set with no
// undo entry, so rewinding an earlier action must not flip it back.
func TestAlwaysAskSurvivesUndo(t *testing.T) {
	g := newActiveGame(t)
	pre := g.Clone()
	if err := g.SetTriggerOrderPreference(g.Seats[0].ID, true); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() { g.RestoreFrom(pre) })
	if !g.Seats[0].TriggerOrderAlwaysAsk {
		t.Error("undo reverted the preference")
	}
}

func TestSetTriggerOrderPreferenceRefusals(t *testing.T) {
	g := NewGame()
	if err := g.SetTriggerOrderPreference(g.ID, true); err == nil {
		t.Error("accepted a preference before the game started")
	}
	g = newActiveGame(t)
	if err := g.SetTriggerOrderPreference(g.ID, true); err == nil {
		t.Error("accepted a preference for a non-seat")
	}
}
