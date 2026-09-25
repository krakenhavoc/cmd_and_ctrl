package game

import (
	"testing"

	"github.com/google/uuid"
)

// advance_step_stack_test.go — #914, CR 117.4.
//
// `advance_step` is the sandbox's skip-ahead button. It used to move
// the turn cursor whatever was on the stack, so a trigger the step
// owed resolved AFTER the next step's turn-based actions: a seat that
// clicked out of declare_blockers took combat damage before the
// afflict trigger, and out of first_strike_damage before the
// first-strike triggers.
//
// CR 117.4 is the rule: a step or phase ends only once every player
// has passed in succession with the stack empty. So the verb means
// "pass priority until this step ends" — the same passes a table of
// humans clicking "next" would make, through the same PassPriority,
// and only then the cursor. These tests pin the four ways that drive
// can end and the one path it must leave alone.

// seedUpkeepWatcher puts a permanent with an upkeep trigger on the
// battlefield. The trigger's body records the step it resolved in and
// the holder's hand size at that moment, which is what "before the
// draw" is measured with. Returns a pointer to the record.
type upkeepWatcherRecord struct {
	resolvedIn    Step
	handOnResolve int
	resolutions   int
}

func seedUpkeepWatcher(t *testing.T, g *Game, owner *Player) *upkeepWatcherRecord {
	t.Helper()
	const oracle = "test-914-upkeep-watcher"
	rec := &upkeepWatcherRecord{}
	g.Battlefield.PushTop(Card{
		InstanceID: uuid.New(),
		Name:       "Test Upkeep Watcher",
		OracleID:   oracle,
		TypeLine:   "Enchantment",
		Owner:      owner.ID,
		Controller: owner.ID,
	})
	withCatalogTriggers(t, func(id string) []TriggeredAbility {
		if id != oracle {
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventBeginUpkeep},
			AppliesTo: func(_ Event, _ *Card, _ Characteristic, _ *Game) bool {
				return true
			},
			Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				return newTriggeredItemForTest(source, "Test Upkeep Watcher — gain 1 life",
					func(g *Game, item *StackItem) error {
						rec.resolutions++
						rec.resolvedIn = g.Turn.Step
						rec.handOnResolve = owner.Hand.Size()
						return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, 1)
					})
			},
		}}
	})
	return rec
}

// raiseUpkeepTrigger fires the watcher and drains it onto the stack,
// leaving the game exactly where a real upkeep leaves it: the trigger
// on the stack, the active player holding priority, nothing resolved.
func raiseUpkeepTrigger(t *testing.T, g *Game) {
	t.Helper()
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventBeginUpkeep, Actor: g.Seats[g.Turn.ActiveSeat].ID})
		g.runStateChecksLocked()
	})
	if len(g.StackMeta) != 1 {
		t.Fatalf("seed: %d items on the stack, want the one upkeep trigger", len(g.StackMeta))
	}
}

// TestAdvanceStepResolvesAnUpkeepTriggerBeforeTheDrawStep is the
// general case behind the combat one: nothing about #914 was specific
// to blocks. An upkeep trigger on the stack resolves IN the upkeep
// step, and the draw step's turn-based draw happens after it —
// CR 117.4 and CR 504.1 in that order.
func TestAdvanceStepResolvesAnUpkeepTriggerBeforeTheDrawStep(t *testing.T) {
	// Four seats, so CR 103.8c has nobody skip turn 1's draw and the
	// draw step has a turn-based action to be late for.
	g := newFourPlayerActiveGame(t)
	active := g.Seats[0]
	rec := seedUpkeepWatcher(t, g, active)
	if g.Turn.Step != StepUpkeep {
		t.Fatalf("a fresh game starts on %s, want upkeep", g.Turn.Step)
	}
	raiseUpkeepTrigger(t, g)
	handBefore := active.Hand.Size()
	lifeBefore := active.Life

	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}

	if rec.resolutions != 1 {
		t.Fatalf("the trigger resolved %d times, want exactly 1", rec.resolutions)
	}
	if rec.resolvedIn != StepUpkeep {
		t.Errorf("the trigger resolved in %s, want upkeep — the step that owed it (CR 117.4)", rec.resolvedIn)
	}
	if rec.handOnResolve != handBefore {
		t.Errorf("hand was %d when the trigger resolved, want %d — the draw step's draw came first",
			rec.handOnResolve, handBefore)
	}
	if active.Life != lifeBefore+1 {
		t.Errorf("life %d, want %d", active.Life, lifeBefore+1)
	}
	if g.Turn.Step != StepDraw {
		t.Errorf("cursor at %s, want draw — the drive ends the step it started in and no further", g.Turn.Step)
	}
	if len(g.StackMeta) != 0 {
		t.Errorf("%d items still on the stack after the step ended", len(g.StackMeta))
	}
	if active.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d, want %d — the draw step still draws", active.Hand.Size(), handBefore+1)
	}
}

// TestAdvanceStepWithAnEmptyStackPassesNoPriority is the path that
// must not move. With nothing owed the verb is what it always was:
// one step of the cursor, and not a single pass. Asserted on the
// drive directly, because a pass round on an empty stack ends the
// step at the same place the cursor move does — the difference is
// visible in what the drive touches, not in where the cursor lands.
func TestAdvanceStepWithAnEmptyStackPassesNoPriority(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	// Hand priority to a seat that is not the active one, so a pass
	// of any kind would be visible in PriorityHolder.
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	holder, step := g.Turn.PriorityHolder, g.Turn.Step
	if holder == g.Turn.ActiveSeat {
		t.Fatalf("setup: priority is still with the active seat %d", holder)
	}

	var got driveResult
	g.WithWriteLock(func() { got = g.driveStepToEndLocked() })

	if got != driveCursorMayMove {
		t.Errorf("drive result %d with an empty stack, want driveCursorMayMove", got)
	}
	if g.Turn.PriorityHolder != holder {
		t.Errorf("priority moved from %d to %d on an empty stack", holder, g.Turn.PriorityHolder)
	}
	if g.Turn.Step != step {
		t.Errorf("the drive moved the cursor to %s; moving it is the caller's job", g.Turn.Step)
	}
}

// TestAdvanceStepWithAnEmptyStackEmitsNothingExtra is the same rule
// measured on the wire: the events an ordinary step advance produces
// are exactly the ones it produced before #914.
func TestAdvanceStepWithAnEmptyStackEmitsNothingExtra(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	before := len(g.Events)

	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}

	var kinds []EventKind
	for _, ev := range g.Events[before:] {
		kinds = append(kinds, ev.Kind)
	}
	want := []EventKind{EventStepBegan}
	if len(kinds) != len(want) {
		t.Fatalf("an empty-stack advance emitted %v, want exactly %v", kinds, want)
	}
	for i := range want {
		if kinds[i] != want[i] {
			t.Fatalf("an empty-stack advance emitted %v, want exactly %v", kinds, want)
		}
	}
}

// TestAdvanceStepStopsAtAPromptTheDriveItselfRaises — the drive holds
// to the same stop autopass does (#730 / ADR 0018 §6): a prompt
// addressed to somebody ends it, with the cursor still in the step
// that owes the rest. It is NOT an error, because the resolutions
// that got this far are real and the caller has to see them; the
// #730 refusal is for a prompt that was already open when the verb
// was called.
func TestAdvanceStepStopsAtAPromptTheDriveItselfRaises(t *testing.T) {
	g := newActiveGame(t)
	active := g.Seats[0]
	const oracle = "test-914-prompting-watcher"
	asked := 0
	g.Battlefield.PushTop(Card{
		InstanceID: uuid.New(),
		Name:       "Test Prompting Watcher",
		OracleID:   oracle,
		TypeLine:   "Enchantment",
		Owner:      active.ID,
		Controller: active.ID,
	})
	withCatalogTriggers(t, func(id string) []TriggeredAbility {
		if id != oracle {
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventBeginUpkeep},
			AppliesTo: func(_ Event, _ *Card, _ Characteristic, _ *Game) bool {
				return true
			},
			Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				return newTriggeredItemForTest(source, "Test Prompting Watcher — scry 1",
					func(g *Game, item *StackItem) error {
						asked++
						g.ScryForEffect(item.Controller, item.SourceCardID, 1)
						return nil
					})
			},
		}}
	})
	raiseUpkeepTrigger(t, g)

	turn, err := g.AdvanceStep()
	if err != nil {
		t.Fatalf("AdvanceStep: %v — a prompt the drive raised is a stop, not a refusal", err)
	}

	if asked != 1 {
		t.Fatalf("the trigger resolved %d times, want 1", asked)
	}
	if turn.Step != StepUpkeep || g.Turn.Step != StepUpkeep {
		t.Errorf("cursor at %s, want upkeep — the drive stopped at the prompt with the step unfinished", g.Turn.Step)
	}
	var prompt *PendingChoice
	g.ReadSnapshot(func() { prompt = g.blockingChoiceLocked() })
	if prompt == nil {
		t.Fatal("no blocking prompt open; the drive stopped for something else")
	}
	if prompt.Kind != PendingChoiceScry {
		t.Errorf("prompt kind %q, want the scry the trigger queued", prompt.Kind)
	}
	// And the #730 refusal is still the answer for a prompt that was
	// already open when the verb was called.
	if _, err := g.AdvanceStep(); err == nil {
		t.Error("a second advance_step with the prompt still open was accepted")
	}
}

// TestAdvanceStepStopsOnALoopNotice — the CR 726 breaker suspends
// AUTOMATIC passing (ADR 0055), and a drive is automatic passing on
// the whole table's behalf. The notice ends it where it stands, so a
// two-permanent loop cannot spin the drive to its bound.
func TestAdvanceStepStopsOnALoopNotice(t *testing.T) {
	const threshold = 3
	g := newActiveGame(t)
	g.LoopThreshold = threshold
	seedTriggerLoop(t, g)
	start := g.Turn

	turn, err := g.AdvanceStep()
	if err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}

	if g.LoopNotice == nil {
		t.Fatal("the drive ran a real loop without the breaker firing")
	}
	if turn.Step != start.Step || turn.Round != start.Round {
		t.Errorf("cursor moved to %s on turn %d; the notice should have stopped the drive at %s",
			turn.Step, turn.Round, start.Step)
	}
	if !g.stackHasItemsLocked() {
		t.Error("the loop's trigger is gone from the stack")
	}
}

// TestUndoAcrossAnAdvanceStepDriveRewindsTheWholeDrive — the drive is
// many resolutions under ONE dispatch, and the room clones before the
// dispatch, so one undo has to take all of it back. Clone / RestoreFrom
// is that seam.
func TestUndoAcrossAnAdvanceStepDriveRewindsTheWholeDrive(t *testing.T) {
	g := newActiveGame(t)
	active := g.Seats[0]
	rec := seedUpkeepWatcher(t, g, active)
	raiseUpkeepTrigger(t, g)
	lifeBefore := active.Life
	handBefore := active.Hand.Size()
	snap := g.Clone()

	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	if active.Life != lifeBefore+1 || g.Turn.Step != StepDraw {
		t.Fatalf("pre-undo: life %d step %s", active.Life, g.Turn.Step)
	}

	g.WithWriteLock(func() { g.RestoreFrom(snap) })

	restored := g.Seats[0]
	if g.Turn.Step != StepUpkeep {
		t.Errorf("undo landed on %s, want upkeep", g.Turn.Step)
	}
	if restored.Life != lifeBefore {
		t.Errorf("undo left the trigger's life gain behind: %d, want %d", restored.Life, lifeBefore)
	}
	if restored.Hand.Size() != handBefore {
		t.Errorf("undo left the draw step's draw behind: %d cards, want %d", restored.Hand.Size(), handBefore)
	}
	if len(g.StackMeta) != 1 {
		t.Fatalf("undo did not put the trigger back on the stack: %d items", len(g.StackMeta))
	}

	// And it replays from there to the same place.
	before := rec.resolutions
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep after undo: %v", err)
	}
	if rec.resolutions != before+1 || restored.Life != lifeBefore+1 || g.Turn.Step != StepDraw {
		t.Errorf("replay after undo: %d resolutions, life %d, step %s",
			rec.resolutions-before, restored.Life, g.Turn.Step)
	}
}
