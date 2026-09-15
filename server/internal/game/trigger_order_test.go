package game

import (
	"testing"

	"github.com/google/uuid"
)

// trigger_order_test.go — S19 sub-PR 8: CR 603.3b same-controller
// trigger ordering. Two differing triggers for one seat hold the
// APNAP drain behind a trigger_order prompt; identical triggers
// don't ask; the chosen order is the resolution order.

// queueTriggerForTest appends a pending triggered item for owner
// whose Effect appends label to *log when it resolves.
func queueTriggerForTest(g *Game, owner *Player, source uuid.UUID, label string, log *[]string) *StackItem {
	item := &StackItem{
		ID:           uuid.New(),
		Kind:         StackItemTriggered,
		Controller:   owner.ID,
		Owner:        owner.ID,
		SourceCardID: source,
		Label:        label,
		Effect: func(_ *Game, it *StackItem) error {
			*log = append(*log, it.Label)
			return nil
		},
	}
	g.PendingTriggers = append(g.PendingTriggers, item)
	return item
}

func findTriggerOrderPrompt(g *Game, chooser uuid.UUID) *PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceTriggerOrder && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

func TestTriggerOrderPromptHoldsDrainUntilAnswered(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	other := g.Seats[2]
	var log []string
	a := queueTriggerForTest(g, me, uuid.New(), "A", &log)
	b := queueTriggerForTest(g, me, uuid.New(), "B", &log)
	c := queueTriggerForTest(g, other, uuid.New(), "C", &log)

	g.WithWriteLock(func() {
		if drained := g.drainPendingTriggersAPNAPLocked(); drained {
			t.Errorf("drain reported drained while an ordering prompt is outstanding")
		}
	})
	if len(g.StackMeta) != 0 {
		t.Fatalf("items reached the stack while seat 0's ordering prompt is pending: %d", len(g.StackMeta))
	}
	if len(g.PendingTriggers) != 3 {
		t.Fatalf("PendingTriggers = %d, want 3 (all held)", len(g.PendingTriggers))
	}
	prompt := findTriggerOrderPrompt(g, me.ID)
	if prompt == nil {
		t.Fatalf("no trigger_order prompt for seat 0")
	}
	if len(prompt.TriggerOrderIDs) != 2 {
		t.Errorf("prompt IDs = %d, want 2", len(prompt.TriggerOrderIDs))
	}
	if findTriggerOrderPrompt(g, other.ID) != nil {
		t.Errorf("seat 2 has a single trigger and must not be asked")
	}
	// Re-running the drain must not duplicate the prompt.
	g.WithWriteLock(func() { g.drainPendingTriggersAPNAPLocked() })
	n := 0
	for _, pc := range g.PendingChoices {
		if pc != nil && pc.Kind == PendingChoiceTriggerOrder {
			n++
		}
	}
	if n != 1 {
		t.Errorf("trigger_order prompts after a second drain: %d, want 1", n)
	}

	// Answer: B resolves before A.
	if err := g.ResolveTriggerOrder(prompt.ID, me.ID, []uuid.UUID{b.ID, a.ID}); err != nil {
		t.Fatalf("ResolveTriggerOrder: %v", err)
	}
	if len(g.PendingTriggers) != 0 {
		t.Fatalf("drain did not run after the answer: %d pending", len(g.PendingTriggers))
	}
	if len(g.StackMeta) != 3 {
		t.Fatalf("stack items = %d, want 3", len(g.StackMeta))
	}
	// Resolve everything: LIFO. APNAP puts the active player's
	// (seat 0) items on first, so seat 2's C resolves first; then B
	// (chosen first), then A.
	g.WithWriteLock(func() {
		for i := 0; i < 3; i++ {
			g.resolveTopAbilityLocked()
		}
	})
	want := []string{"C", "B", "A"}
	if len(log) != 3 || log[0] != want[0] || log[1] != want[1] || log[2] != want[2] {
		t.Errorf("resolution order = %v, want %v", log, want)
	}
	_ = c
}

func TestTriggerOrderIdenticalTriggersDoNotPrompt(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	var log []string
	src := uuid.New()
	queueTriggerForTest(g, me, src, "Bident of Thassa — draw a card", &log)
	queueTriggerForTest(g, me, src, "Bident of Thassa — draw a card", &log)
	g.WithWriteLock(func() {
		if !g.drainPendingTriggersAPNAPLocked() {
			t.Errorf("identical triggers should drain without a prompt")
		}
	})
	if findTriggerOrderPrompt(g, me.ID) != nil {
		t.Errorf("prompt queued for two interchangeable triggers")
	}
	if len(g.StackMeta) != 2 {
		t.Errorf("stack items = %d, want 2", len(g.StackMeta))
	}
}

func TestTriggerOrderRejectsBadPermutationAndWrongChooser(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	var log []string
	a := queueTriggerForTest(g, me, uuid.New(), "A", &log)
	queueTriggerForTest(g, me, uuid.New(), "B", &log)
	g.WithWriteLock(func() { g.drainPendingTriggersAPNAPLocked() })
	prompt := findTriggerOrderPrompt(g, me.ID)
	if prompt == nil {
		t.Fatalf("no prompt")
	}
	if err := g.ResolveTriggerOrder(prompt.ID, g.Seats[1].ID, []uuid.UUID{a.ID, prompt.TriggerOrderIDs[1]}); err != ErrNotTheChooser {
		t.Errorf("wrong chooser: got %v, want ErrNotTheChooser", err)
	}
	if err := g.ResolveTriggerOrder(prompt.ID, me.ID, []uuid.UUID{a.ID, a.ID}); err != ErrInvalidParam {
		t.Errorf("duplicate IDs: got %v, want ErrInvalidParam", err)
	}
	if err := g.ResolveTriggerOrder(prompt.ID, me.ID, []uuid.UUID{a.ID}); err != ErrInvalidParam {
		t.Errorf("short list: got %v, want ErrInvalidParam", err)
	}
	if len(g.PendingTriggers) != 2 || len(g.StackMeta) != 0 {
		t.Errorf("rejected answers must leave the queue held")
	}
}

func TestTriggerOrderLateArrivalReprompts(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	var log []string
	a := queueTriggerForTest(g, me, uuid.New(), "A", &log)
	b := queueTriggerForTest(g, me, uuid.New(), "B", &log)
	g.WithWriteLock(func() { g.drainPendingTriggersAPNAPLocked() })
	prompt := findTriggerOrderPrompt(g, me.ID)
	// A third, different trigger shows up while the prompt is open.
	queueTriggerForTest(g, me, uuid.New(), "C", &log)
	if err := g.ResolveTriggerOrder(prompt.ID, me.ID, []uuid.UUID{a.ID, b.ID}); err != nil {
		t.Fatalf("ResolveTriggerOrder: %v", err)
	}
	// A and B are Ordered, C is not → asked again with all three.
	second := findTriggerOrderPrompt(g, me.ID)
	if second == nil {
		t.Fatalf("no re-prompt after a late-arriving trigger")
	}
	if len(second.TriggerOrderIDs) != 3 {
		t.Errorf("re-prompt IDs = %d, want 3", len(second.TriggerOrderIDs))
	}
	if len(g.StackMeta) != 0 {
		t.Errorf("nothing should reach the stack until the re-prompt is answered")
	}
}

func TestTriggerOrderSurvivesClone(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	var log []string
	queueTriggerForTest(g, me, uuid.New(), "A", &log)
	queueTriggerForTest(g, me, uuid.New(), "B", &log)
	g.WithWriteLock(func() { g.drainPendingTriggersAPNAPLocked() })
	snap := g.Clone()
	p := findTriggerOrderPrompt(snap, me.ID)
	if p == nil || len(p.TriggerOrderIDs) != 2 {
		t.Fatalf("cloned prompt lost its IDs: %+v", p)
	}
	if len(snap.PendingTriggers) != 2 {
		t.Errorf("cloned PendingTriggers = %d, want 2", len(snap.PendingTriggers))
	}
}
