package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// event_batch_test.go — #829. "Whenever ONE OR MORE …" collapses an
// EVENT BATCH and nothing wider: the same batch fires the ability
// once, and the next batch fires it again however much of the last
// one is still on the stack or waiting on a prompt (CR 603.2c).
//
// #594's own cases are here too (one batch, many events, one
// trigger — with and without an optional prompt), because they are
// the half that must not move.

const (
	batchProbeOracleID = "test-event-batch-probe"
	batchProbeLabel    = "Batch Probe — trigger"
)

// seedBatchProbe puts a permanent on the battlefield whose single
// ability is OncePerBatch and watches EventETB for any card but
// itself, and points CatalogTriggers at it for the test. `optional`
// gives the ability a "you may" prompt, which is the shape #594's
// prompt-window case used.
func seedBatchProbe(t *testing.T, g *Game, optional bool) uuid.UUID {
	t.Helper()
	owner := g.Seats[0]
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id,
		Name:       "Batch Probe",
		OracleID:   batchProbeOracleID,
		TypeLine:   "Enchantment",
		Owner:      owner.ID,
		Controller: owner.ID,
	})
	ability := TriggeredAbility{
		Watches:      []EventKind{EventETB},
		Key:          batchProbeLabel,
		OncePerBatch: true,
		AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
			return ev.CardID != source.InstanceID
		},
		Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			return NewTriggeredItem(source, batchProbeLabel, func(*Game, *StackItem) error { return nil })
		},
	}
	if optional {
		ability.OptionalPrompt = &TriggerOptionalPrompt{Question: "Batch Probe?"}
	}
	withCatalogTriggers(t, func(oracle string) []TriggeredAbility {
		if oracle != batchProbeOracleID {
			return nil
		}
		return []TriggeredAbility{ability}
	})
	return id
}

// emitProbeEvent emits one EventETB for a card that is not on the
// battlefield — enough to match the probe's AppliesTo without
// dragging entry plumbing into a test about batching.
func emitProbeEvent(g *Game) {
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventETB, Actor: g.Seats[0].ID, CardID: uuid.New()})
	})
}

// probeItems counts the probe's triggered items wherever they are:
// waiting on PendingTriggers or already on the stack.
func probeItems(g *Game, source uuid.UUID) int {
	n := 0
	for _, it := range g.PendingTriggers {
		if it != nil && it.SourceCardID == source {
			n++
		}
	}
	for _, it := range g.StackMeta {
		if it != nil && it.Kind == StackItemTriggered && it.SourceCardID == source {
			n++
		}
	}
	return n
}

// probePrompts counts the probe's unanswered "you may" prompts.
func probePrompts(g *Game, source uuid.UUID) int {
	n := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceTriggerPrompt && c.Source == source {
			n++
		}
	}
	return n
}

func lastEventBatch(g *Game) uint64 {
	if len(g.Events) == 0 {
		return 0
	}
	return g.Events[len(g.Events)-1].Batch
}

// TestOncePerBatchCollapsesOneBatchOnly is #594's case and #829's,
// side by side: three events of one batch are one trigger, and one
// event of the NEXT batch is a second trigger even though the first
// is still queued.
func TestOncePerBatchCollapsesOneBatchOnly(t *testing.T) {
	g := newActiveGame(t)
	probe := seedBatchProbe(t, g, false)

	g.WithWriteLock(func() { g.beginEventBatchLocked() })
	for i := 0; i < 3; i++ {
		emitProbeEvent(g)
	}
	if n := probeItems(g, probe); n != 1 {
		t.Fatalf("%d triggers for three events of one batch, want 1", n)
	}

	g.WithWriteLock(func() { g.beginEventBatchLocked() })
	emitProbeEvent(g)
	if n := probeItems(g, probe); n != 2 {
		t.Fatalf("%d triggers after a second batch, want 2 — a new batch triggers again", n)
	}
}

// TestOncePerBatchIgnoresAnItemAlreadyOnTheStack is the reported bug:
// Dour Port-Mage's shape, where the first batch's trigger has left
// PendingTriggers and is sitting on the stack when the second batch
// arrives.
func TestOncePerBatchIgnoresAnItemAlreadyOnTheStack(t *testing.T) {
	g := newActiveGame(t)
	probe := seedBatchProbe(t, g, false)

	g.WithWriteLock(func() { g.beginEventBatchLocked() })
	emitProbeEvent(g)
	g.WithWriteLock(func() { g.drainPendingTriggersAPNAPLocked() })
	if len(g.PendingTriggers) != 0 {
		t.Fatalf("%d triggers still pending, want the drain to have put it on the stack", len(g.PendingTriggers))
	}
	if n := probeItems(g, probe); n != 1 {
		t.Fatalf("%d triggers on the stack after the first batch, want 1", n)
	}

	g.WithWriteLock(func() { g.beginEventBatchLocked() })
	emitProbeEvent(g)
	if n := probeItems(g, probe); n != 2 {
		t.Errorf("%d triggers, want 2 — the first batch's item on the stack must not swallow a second batch", n)
	}
}

// TestOncePerBatchIgnoresAnUnansweredPrompt is the same, one step
// earlier: the first batch's trigger has not even been built yet
// because its "you may" prompt is unanswered.
//
// A trigger prompt blocks the table (choice_gate.go), so no real
// action can reach a second batch while one is open — this drives the
// boundary directly, which is what makes the assertion about the
// GUARD rather than about who may act.
func TestOncePerBatchIgnoresAnUnansweredPrompt(t *testing.T) {
	g := newActiveGame(t)
	probe := seedBatchProbe(t, g, true)

	g.WithWriteLock(func() { g.beginEventBatchLocked() })
	emitProbeEvent(g)
	emitProbeEvent(g)
	if n := probePrompts(g, probe); n != 1 {
		t.Fatalf("%d prompts for two events of one batch, want 1 (#594)", n)
	}

	g.WithWriteLock(func() { g.beginEventBatchLocked() })
	emitProbeEvent(g)
	if n := probePrompts(g, probe); n != 2 {
		t.Errorf("%d prompts, want 2 — an unanswered prompt must not swallow a second batch", n)
	}
}

// TestEventBatchAdvancesAtStepEntryAndResolutionOnly pins the
// boundary itself, which is the thing a card author has to be able to
// predict. Two events with nothing resolving and no step change in
// between are ONE batch — that is what keeps a three-creature attack
// declared one DeclareAttacker at a time to a single Adeline trigger.
func TestEventBatchAdvancesAtStepEntryAndResolutionOnly(t *testing.T) {
	g := newActiveGame(t)
	probe := seedBatchProbe(t, g, false)

	emitProbeEvent(g)
	first := lastEventBatch(g)
	if first == 0 {
		t.Fatal("an emitted event carries no batch")
	}
	emitProbeEvent(g)
	if got := lastEventBatch(g); got != first {
		t.Errorf("batch %d then %d — two events with nothing resolving and no step change are one batch", first, got)
	}
	if n := probeItems(g, probe); n != 1 {
		t.Fatalf("%d triggers for that one batch, want 1", n)
	}

	// A step entry is a boundary.
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	emitProbeEvent(g)
	afterStep := lastEventBatch(g)
	if afterStep <= first {
		t.Errorf("batch %d after a step entry, want greater than %d", afterStep, first)
	}

	// A resolution is the other boundary, and it is not a step
	// change: the cursor must still be where it was.
	step := g.Turn.Step
	for i := 0; i < 16 && !stackEmptyForBatchTest(g); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if g.Turn.Step != step {
		t.Fatalf("the step advanced to %q while the stack drained; this case is about resolutions", g.Turn.Step)
	}
	emitProbeEvent(g)
	if got := lastEventBatch(g); got <= afterStep {
		t.Errorf("batch %d after a resolution, want greater than %d", got, afterStep)
	}
}

func stackEmptyForBatchTest(g *Game) bool {
	if g.Stack != nil && g.Stack.Size() > 0 {
		return false
	}
	if len(g.PendingTriggers) > 0 {
		return false
	}
	for _, it := range g.StackMeta {
		if it != nil {
			return false
		}
	}
	return true
}

// TestEventBatchSurvivesCloneAndRestore — an undo rewinds the counter
// and the once-per-batch marks together, so redoing the undone action
// neither double-fires the trigger nor loses it.
func TestEventBatchSurvivesCloneAndRestore(t *testing.T) {
	g := newActiveGame(t)
	probe := seedBatchProbe(t, g, false)

	g.WithWriteLock(func() { g.beginEventBatchLocked() })
	pre := g.Clone()

	emitProbeEvent(g)
	emitProbeEvent(g)
	if n := probeItems(g, probe); n != 1 {
		t.Fatalf("%d triggers before the undo, want 1", n)
	}

	g.WithWriteLock(func() { g.RestoreFrom(pre) })
	if n := probeItems(g, probe); n != 0 {
		t.Fatalf("%d triggers survived the undo, want 0", n)
	}

	// Redo: the batch is the same one, so the trigger fires once for
	// it — not zero (swallowed by a mark the undo failed to rewind)
	// and not twice (a counter that rewound without its marks).
	emitProbeEvent(g)
	if n := probeItems(g, probe); n != 1 {
		t.Fatalf("%d triggers after redoing the batch, want 1", n)
	}
	emitProbeEvent(g)
	if n := probeItems(g, probe); n != 1 {
		t.Errorf("%d triggers for the redone batch, want 1", n)
	}
}

// TestEventBatchLeavesTheTurnTallyAlone — the stamp is additive: the
// per-turn tally and EventsThisTurn count exactly what they counted
// before, and every event carries a batch.
func TestEventBatchLeavesTheTurnTallyAlone(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	beforeEvents := len(g.EventsThisTurn())
	beforeDraws := g.TurnTallyFor(me.ID).CardsDrawn

	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventDrawCard, Actor: me.ID, CardID: uuid.New()})
		g.EmitEvent(Event{Kind: EventDrawCard, Actor: me.ID, CardID: uuid.New()})
	})

	turn := g.EventsThisTurn()
	if got := len(turn) - beforeEvents; got != 2 {
		t.Errorf("EventsThisTurn grew by %d, want 2", got)
	}
	if got := g.TurnTallyFor(me.ID).CardsDrawn - beforeDraws; got != 2 {
		t.Errorf("CardsDrawn grew by %d, want 2", got)
	}
	for _, ev := range turn {
		if ev.Batch == 0 {
			t.Fatalf("event seq %d carries no batch", ev.Seq)
		}
	}
	last := turn[len(turn)-1]
	if last.Batch != turn[len(turn)-2].Batch {
		t.Error("two events emitted in one mutation with nothing resolving must share a batch")
	}
}

// TestEventBatchSurvivesThePersistedSnapshot — a restart mid-batch
// comes back with the same batch open and the same marks against it,
// so an ability does not fire a second time for a batch it has
// already fired for.
func TestEventBatchSurvivesThePersistedSnapshot(t *testing.T) {
	g := newActiveGame(t)
	probe := seedBatchProbe(t, g, false)

	g.WithWriteLock(func() { g.beginEventBatchLocked() })
	emitProbeEvent(g)
	if n := probeItems(g, probe); n != 1 {
		t.Fatalf("%d triggers before the snapshot, want 1", n)
	}
	batch := lastEventBatch(g)

	blob, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}
	var decoded GameSnapshot
	if err := json.Unmarshal(blob, &decoded); err != nil {
		t.Fatalf("unmarshal snapshot: %v", err)
	}
	restored, err := decoded.Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}

	restored.WithWriteLock(func() {
		restored.EmitEvent(Event{Kind: EventETB, Actor: restored.Seats[0].ID, CardID: uuid.New()})
	})
	if got := lastEventBatch(restored); got != batch {
		t.Errorf("batch %d after the restore, want the open batch %d", got, batch)
	}
	if n := probeItems(restored, probe); n != 1 {
		t.Errorf("%d triggers after the restore, want 1 — the batch's mark must survive the restart", n)
	}
}
