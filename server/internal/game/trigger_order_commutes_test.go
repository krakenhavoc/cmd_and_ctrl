package game

import (
	"testing"

	"github.com/google/uuid"
)

// trigger_order_commutes_test.go — #1511: a seat whose whole batch is
// StackItem.Commutes items (prowess instances on different creatures)
// drains without a CR 603.3b ordering prompt; one item outside the
// class, or one that picked up a target or a mode, brings it back.

// queueProwessForTest appends the item the real prowess trigger
// builds for a fresh prowess creature of owner's.
func queueProwessForTest(g *Game, owner *Player) *StackItem {
	src := Card{InstanceID: uuid.New(), Name: "Monk", TypeLine: "Creature — Monk", Owner: owner.ID, Controller: owner.ID}
	item := prowessTrigger.Build(Event{}, &src, Characteristic{}, g)
	item.ID = uuid.New()
	g.PendingTriggers = append(g.PendingTriggers, item)
	return item
}

func TestProwessTriggerIsBuiltCommuting(t *testing.T) {
	g := newActiveGame(t)
	item := queueProwessForTest(g, g.Seats[0])
	if !item.Commutes {
		t.Fatal("the prowess trigger's Build no longer marks its item Commutes (#1511)")
	}
	if len(item.Targets) != 0 || len(item.Modes) != 0 {
		t.Fatalf("a prowess item carries targets %v / modes %v", item.Targets, item.Modes)
	}
}

func TestSeatNeedsTriggerOrderCommutingBatches(t *testing.T) {
	commuting := func() *StackItem {
		return &StackItem{ID: uuid.New(), SourceCardID: uuid.New(), Label: prowessLabel, Commutes: true}
	}
	other := func(label string) *StackItem {
		return &StackItem{ID: uuid.New(), SourceCardID: uuid.New(), Label: label}
	}
	targeted := commuting()
	targeted.Targets = []TargetRef{{Kind: TargetPlayer, ID: uuid.New()}}
	modal := commuting()
	modal.Modes = []int{0}

	cases := []struct {
		name  string
		items []*StackItem
		want  bool
	}{
		{"two commuting, different sources", []*StackItem{commuting(), commuting()}, false},
		{"three commuting, different sources", []*StackItem{commuting(), commuting(), commuting()}, false},
		{"commuting plus one untargeted other", []*StackItem{commuting(), commuting(), other("Mentor — make a Monk")}, true},
		{"commuting plus one that picked a target", []*StackItem{commuting(), targeted}, true},
		{"commuting plus one that picked a mode", []*StackItem{commuting(), modal}, true},
		{"two differing non-commuting", []*StackItem{other("A"), other("B")}, true},
		{"one item", []*StackItem{other("A")}, false},
	}
	for _, c := range cases {
		if got := seatNeedsTriggerOrder(c.items); got != c.want {
			t.Errorf("%s: seatNeedsTriggerOrder = %v, want %v", c.name, got, c.want)
		}
	}
}

// TestAllProwessBatchDrainsWithoutAPrompt — the real drain, the real
// prowess Build: three creatures' triggers go straight onto the stack.
func TestAllProwessBatchDrainsWithoutAPrompt(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	for i := 0; i < 3; i++ {
		queueProwessForTest(g, me)
	}
	g.WithWriteLock(func() {
		if !g.drainPendingTriggersAPNAPLocked() {
			t.Error("an all-prowess batch was held behind an ordering prompt")
		}
	})
	if findTriggerOrderPrompt(g, me.ID) != nil {
		t.Error("an all-prowess batch queued a trigger_order prompt")
	}
	if len(g.StackMeta) != 3 {
		t.Errorf("stack items = %d, want 3", len(g.StackMeta))
	}
}

// TestProwessBatchWithATargetedTriggerStillPrompts — the same batch
// plus one targeted trigger is held, and the prompt lists all four.
func TestProwessBatchWithATargetedTriggerStillPrompts(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	for i := 0; i < 3; i++ {
		queueProwessForTest(g, me)
	}
	var log []string
	burn := queueTriggerForTest(g, me, uuid.New(), "Caldera Pyremaw — damage equal to its power to target opponent", &log)
	burn.Targets = []TargetRef{{Kind: TargetPlayer, ID: g.Seats[1].ID}}

	g.WithWriteLock(func() {
		if g.drainPendingTriggersAPNAPLocked() {
			t.Error("prowess plus a targeted trigger drained without an ordering prompt")
		}
	})
	prompt := findTriggerOrderPrompt(g, me.ID)
	if prompt == nil {
		t.Fatal("no trigger_order prompt for prowess plus a targeted trigger")
	}
	if len(prompt.TriggerOrderIDs) != 4 {
		t.Errorf("prompt orders %d items, want 4", len(prompt.TriggerOrderIDs))
	}
	if len(g.StackMeta) != 0 {
		t.Errorf("%d items reached the stack while the prompt is open", len(g.StackMeta))
	}
}

// TestCommutesSurvivesCloneAndRestore — undo is Clone + RestoreFrom,
// and a restored batch must drain exactly as the original would.
func TestCommutesSurvivesCloneAndRestore(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	queueProwessForTest(g, me)
	queueProwessForTest(g, me)

	snap := g.Clone()
	g.WithWriteLock(func() { g.drainPendingTriggersAPNAPLocked() })
	if len(g.StackMeta) != 2 {
		t.Fatalf("stack items = %d before undo, want 2", len(g.StackMeta))
	}

	g.RestoreFrom(snap)
	if len(g.PendingTriggers) != 2 || len(g.StackMeta) != 0 {
		t.Fatalf("restore: %d pending, %d on the stack; want 2 and 0", len(g.PendingTriggers), len(g.StackMeta))
	}
	for _, it := range g.PendingTriggers {
		if !it.Commutes {
			t.Fatalf("the restored item %s lost Commutes", it.ID)
		}
	}
	g.WithWriteLock(func() {
		if !g.drainPendingTriggersAPNAPLocked() {
			t.Error("the restored all-prowess batch was held behind a prompt")
		}
	})
	if findTriggerOrderPrompt(g, me.ID) != nil {
		t.Error("the restored all-prowess batch queued a trigger_order prompt")
	}
}
