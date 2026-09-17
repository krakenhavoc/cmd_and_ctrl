package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// eliminated_chooser_test.go — #864: a pending choice whose chooser
// has left the game must never sit in Game.PendingChoices, in either
// direction of the race:
//
//   - the chooser is ALREADY eliminated at the moment something tries
//     to queue a choice for them (QueueChoiceForEffect's own guard,
//     plus the two callers that had to be taught not to treat a
//     refused queue as a still-open prompt);
//   - the chooser was alive when their choice was queued and is
//     eliminated later, while it sits open (sweepEliminatedChoicesLocked,
//     run at every runStateChecksLocked pass as a backstop behind the
//     #287 sweep that already runs at the instant of elimination).

// TestQueueChoiceForEffectRefusesEliminatedChooser is the direct,
// per-kind check on the queue-time guard: none of the three kinds the
// S31 bot arena caught wedging (pick_target, trigger_prompt,
// trigger_order) can be queued for a chooser who has already left.
func TestQueueChoiceForEffectRefusesEliminatedChooser(t *testing.T) {
	kinds := []PendingChoiceKind{
		PendingChoicePickTarget,
		PendingChoiceTriggerPrompt,
		PendingChoiceTriggerOrder,
	}
	for _, kind := range kinds {
		t.Run(string(kind), func(t *testing.T) {
			g := newActiveGame(t)
			gone := g.Seats[1]
			source := uuid.New()
			var id uuid.UUID
			var before int
			g.WithWriteLock(func() {
				gone.Eliminated = true
				before = len(g.Events)
				id = g.QueueChoiceForEffect(PendingChoice{
					Kind:    kind,
					Chooser: gone.ID,
					Count:   1,
					Source:  source,
					Reason:  "test",
				})
			})
			if id != uuid.Nil {
				t.Errorf("QueueChoiceForEffect returned %s, want uuid.Nil", id)
			}
			if len(g.PendingChoices) != 0 {
				t.Fatalf("PendingChoices = %d, want 0: %+v", len(g.PendingChoices), g.PendingChoices)
			}
			var dropped *Event
			for i := before; i < len(g.Events); i++ {
				if g.Events[i].Kind == EventPendingChoiceDropped {
					dropped = &g.Events[i]
					break
				}
			}
			if dropped == nil {
				t.Fatalf("no EventPendingChoiceDropped emitted")
			}
			if dropped.Actor != gone.ID {
				t.Errorf("dropped event Actor = %s, want %s", dropped.Actor, gone.ID)
			}
			if dropped.Label != string(kind) {
				t.Errorf("dropped event Label = %q, want %q", dropped.Label, kind)
			}
		})
	}
}

// TestQueueChoiceForEffectRefusesUnseatedChooser covers the other
// half of the guard's condition: a chooser who was never seated at
// all (defensive — nothing in this codebase queues for a stranger,
// but the guard's "p == nil" arm needs its own coverage).
func TestQueueChoiceForEffectRefusesUnseatedChooser(t *testing.T) {
	g := newActiveGame(t)
	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.QueueChoiceForEffect(PendingChoice{
			Kind:    PendingChoiceTriggerPrompt,
			Chooser: uuid.New(),
			Count:   1,
			Reason:  "test",
		})
	})
	if id != uuid.Nil {
		t.Errorf("QueueChoiceForEffect returned %s, want uuid.Nil", id)
	}
	if len(g.PendingChoices) != 0 {
		t.Fatalf("PendingChoices = %d, want 0", len(g.PendingChoices))
	}
}

// TestDispatchTriggerDropsTriggerPromptForEliminatedController drives
// the REAL CR 603.5 dispatch path (dispatchTriggerLocked) for a
// Solemn-Simulacrum-shaped optional trigger whose controller has
// already left — the exact shape of the original report.
func TestDispatchTriggerDropsTriggerPromptForEliminatedController(t *testing.T) {
	g := newActiveGame(t)
	gone := g.Seats[1]
	source := Card{InstanceID: uuid.New(), Name: "Solemn Simulacrum", Controller: gone.ID, Owner: gone.ID}
	ability := TriggeredAbility{
		OptionalPrompt: &TriggerOptionalPrompt{Question: "Draw a card?"},
		Build: func(ev Event, source *Card, lki Characteristic, g *Game) *StackItem {
			t.Fatalf("Build ran for an eliminated controller's optional trigger")
			return nil
		},
	}
	g.WithWriteLock(func() {
		gone.Eliminated = true
		g.dispatchTriggerLocked(Event{Kind: EventLTB, CardID: source.InstanceID}, source, source.Effective(), ability)
	})
	if len(g.PendingChoices) != 0 {
		t.Fatalf("PendingChoices = %d, want 0: %+v", len(g.PendingChoices), g.PendingChoices)
	}
}

// TestDispatchTriggerDropsPickTargetForEliminatedController drives the
// REAL CR 603.3d target-pick dispatch (dispatchTriggerLocked ->
// queuePickTargetLocked) for a targeted trigger whose controller has
// already left.
func TestDispatchTriggerDropsPickTargetForEliminatedController(t *testing.T) {
	g := newActiveGame(t)
	gone := g.Seats[1]
	source := Card{InstanceID: uuid.New(), Name: "Test Targeter", Controller: gone.ID, Owner: gone.ID}
	ability := TriggeredAbility{
		Targets: &TargetSpec{Mode: "any", Label: "any target", Players: true, Min: 1, Max: 1},
		Build: func(ev Event, source *Card, lki Characteristic, g *Game) *StackItem {
			t.Fatalf("Build ran for an eliminated controller's targeted trigger")
			return nil
		},
	}
	g.WithWriteLock(func() {
		gone.Eliminated = true
		g.dispatchTriggerLocked(Event{Kind: EventLTB, CardID: source.InstanceID}, source, source.Effective(), ability)
	})
	if len(g.PendingChoices) != 0 {
		t.Fatalf("PendingChoices = %d, want 0: %+v", len(g.PendingChoices), g.PendingChoices)
	}
}

// TestDrainPendingTriggersDropsOrderForEliminatedController drives the
// REAL CR 603.3b drain (drainPendingTriggersAPNAPLocked) for a seat
// with two differing simultaneous triggers who has since left. Before
// #864 this held the ENTIRE APNAP drain forever: the seat's own
// ordering prompt could never be created (QueueChoiceForEffect's
// guard refuses it) or answered, but the surrounding loop still
// marked itself `held` and never re-checked. This is the caller #864
// singles out as one the generic guard alone does not fix.
func TestDrainPendingTriggersDropsOrderForEliminatedController(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	gone := g.Seats[0]
	live := g.Seats[2]
	var log []string
	queueTriggerForTest(g, gone, uuid.New(), "A", &log)
	queueTriggerForTest(g, gone, uuid.New(), "B", &log)
	queueTriggerForTest(g, live, uuid.New(), "C", &log)

	var drained bool
	g.WithWriteLock(func() {
		gone.Eliminated = true
		drained = g.drainPendingTriggersAPNAPLocked()
	})
	if !drained {
		t.Fatalf("drain reported held, want fully drained (eliminated seat must not hold the table)")
	}
	if findTriggerOrderPrompt(g, gone.ID) != nil {
		t.Errorf("trigger_order prompt queued for an eliminated chooser")
	}
	if len(g.PendingTriggers) != 0 {
		t.Fatalf("PendingTriggers = %d, want 0 (drained or dropped)", len(g.PendingTriggers))
	}
	// gone's two triggers are dropped outright (CR 800.4a — their
	// controller left, so the abilities cease to exist); live's
	// single trigger still resolves normally.
	if len(g.StackMeta) != 1 {
		t.Fatalf("StackMeta = %d, want 1 (only live's trigger)", len(g.StackMeta))
	}
	for _, item := range g.StackMeta {
		if item.Controller != live.ID {
			t.Errorf("surviving stack item controller = %s, want %s", item.Controller, live.ID)
		}
	}
}

// TestSweepEliminatedChoicesDropsChoiceForNewlyEliminatedChooser
// covers the belt-and-braces half: a choice queued while its chooser
// was alive, left pending when they are eliminated by some route
// other than leaveGameLocked's own instantaneous #287 sweep (that
// existing sweep is deliberately bypassed here — flipping Eliminated
// directly — so this test exercises ONLY the new
// sweepEliminatedChoicesLocked backstop, not the old one). Also
// checks the actual symptom from #864: the table must not be left
// blocked behind the dropped choice.
func TestSweepEliminatedChoicesDropsChoiceForNewlyEliminatedChooser(t *testing.T) {
	g := newActiveGame(t)
	chooser := g.Seats[1]
	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.QueueChoiceForEffect(PendingChoice{
			Kind:    PendingChoiceTriggerPrompt,
			Chooser: chooser.ID,
			Count:   1,
			Reason:  "test",
		})
	})
	if id == uuid.Nil {
		t.Fatalf("setup: choice was not queued for a live chooser")
	}
	if err := g.PassPriority(); !errors.Is(err, ErrChoicePending) {
		t.Fatalf("PassPriority before elimination = %v, want ErrChoicePending", err)
	}

	g.WithWriteLock(func() {
		chooser.Eliminated = true // NOT via leaveGameLocked — see doc comment.
		g.runStateChecksLocked()
	})

	if len(g.PendingChoices) != 0 {
		t.Fatalf("PendingChoices = %d, want 0 after the sweep: %+v", len(g.PendingChoices), g.PendingChoices)
	}
	if err := g.PassPriority(); errors.Is(err, ErrChoicePending) {
		t.Errorf("PassPriority after the sweep still refused on the dropped choice: %v", err)
	} else if err != nil {
		t.Errorf("PassPriority after the sweep: unexpected error %v", err)
	}

	// Idempotent: running it again over an already-clean queue must
	// not panic or misbehave.
	g.WithWriteLock(func() { g.sweepEliminatedChoicesLocked() })
}

// TestSweepEliminatedChoicesLeavesLiveChoosersChoice is the negative
// case the doc comment promises: the sweep must never touch a choice
// whose chooser is still in the game.
func TestSweepEliminatedChoicesLeavesLiveChoosersChoice(t *testing.T) {
	g := newActiveGame(t)
	live := g.Seats[1]
	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.QueueChoiceForEffect(PendingChoice{
			Kind:    PendingChoiceTriggerPrompt,
			Chooser: live.ID,
			Count:   1,
			Reason:  "test",
		})
		g.sweepEliminatedChoicesLocked()
	})
	if id == uuid.Nil {
		t.Fatalf("setup: choice was not queued")
	}
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].ID != id {
		t.Fatalf("live chooser's choice was dropped by the sweep: %+v", g.PendingChoices)
	}
}

// TestEliminatedDuringOwnResolutionReachesDecidedState is the
// regression test in the shape of the real failure the S31 bot arena
// found (#864): a player's OWN ability is resolving when a
// state-based action eliminates them mid-resolution — the same
// window Solemn Simulacrum's death trigger and its controller's life
// hitting 0 shared — and the resolution goes on to queue a further
// choice in that now-departed player's name. The choice must be
// refused, not left pending, and (two players, one left) the game
// must reach StateEnded rather than sit blocked forever.
func TestEliminatedDuringOwnResolutionReachesDecidedState(t *testing.T) {
	g := newActiveGame(t)
	dying := g.Seats[0]
	dying.Life = 1

	item := &StackItem{
		ID:           uuid.New(),
		Kind:         StackItemTriggered,
		Controller:   dying.ID,
		Owner:        dying.ID,
		SourceCardID: uuid.New(),
		Label:        "test — dies while its own ability resolves",
		Effect: func(g *Game, it *StackItem) error {
			// The ability's own resolution knocks its controller to 0
			// life (standing in for "a state-based action fires in the
			// same window this ability resolves in" — #864's reported
			// sequence).
			dying.Life = 0
			g.runStateChecksLocked()
			// Something the resolving ability's own harvester chain
			// still tries to ask ITS controller, after that controller
			// has already left. Must be refused, not queued.
			id := g.QueueChoiceForEffect(PendingChoice{
				Kind:    PendingChoiceTriggerPrompt,
				Chooser: dying.ID,
				Count:   1,
				Reason:  "may draw a card",
			})
			if id != uuid.Nil {
				t.Errorf("choice queued for a chooser eliminated earlier in this same resolution")
			}
			return nil
		},
	}

	g.WithWriteLock(func() {
		if g.StackMeta == nil {
			g.StackMeta = make(map[uuid.UUID]*StackItem)
		}
		item.Seq = g.nextStackSeqLocked()
		g.StackMeta[item.ID] = item
		g.resolveTopAbilityLocked()
	})

	if len(g.PendingChoices) != 0 {
		t.Fatalf("PendingChoices = %d, want 0: %+v", len(g.PendingChoices), g.PendingChoices)
	}
	if !dying.Eliminated {
		t.Fatalf("dying player was never marked eliminated")
	}
	if g.State != StateEnded {
		t.Fatalf("game state = %v, want %v (decided, not wedged)", g.State, StateEnded)
	}
}
