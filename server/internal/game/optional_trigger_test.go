package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// optional_trigger_test.go — CR 603.5 (#2494): an optional trigger goes
// on the stack like any other and asks its "you may" as it resolves.

// passUntilTriggerPrompt passes priority until a trigger_prompt is open
// and returns it, or fails when the stack empties first.
func passUntilTriggerPrompt(t *testing.T, g *Game) *PendingChoice {
	t.Helper()
	for i := 0; i < 32; i++ {
		if c := pendingChoiceOfKind(g, PendingChoiceTriggerPrompt); c != nil {
			return c
		}
		if g.Stack.Size() == 0 && len(g.StackMeta) == 0 && len(g.PendingTriggers) == 0 {
			break
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority iter %d: %v", i, err)
		}
	}
	t.Fatal("no trigger_prompt before the stack emptied")
	return nil
}

// scheduleOptionalOnCast queues an optional "when you next cast a
// spell this turn, you may …" whose effect bumps *ran. Its effect is a
// registered body, so the item is plain data and a snapshot of it is a
// restore point.
func scheduleOptionalOnCast(g *Game, controller uuid.UUID, ran *int) {
	g.WithWriteLock(func() {
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller:       controller,
			Label:            "probe — you may draw",
			On:               []EventKind{EventCast},
			OptionalQuestion: "Draw a card?",
			Condition: testCondition(func(ev Event, dt *DelayedTrigger, _ *Game) bool {
				return ev.Actor == dt.Controller
			}),
			Body: testBody(func(*Game, *StackItem) error { *ran++; return nil }),
		})
	})
}

// optionalItemOnStack is the one optional triggered item in StackMeta,
// or nil.
func optionalItemOnStack(g *Game) *StackItem {
	for _, it := range g.StackMeta {
		if it != nil && it.Kind == StackItemTriggered && it.Optional != nil {
			return it
		}
	}
	return nil
}

func TestOptionalTriggerGoesOnTheStackAndAsksAsItResolves(t *testing.T) {
	for _, tc := range []struct {
		name   string
		answer bool
		want   int
	}{{"yes runs the effect", true, 1}, {"no resolves doing nothing", false, 0}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			me := g.Seats[0]
			ran := 0
			scheduleOptionalOnCast(g, me.ID, &ran)

			emitCast(g, me.ID, uuid.New())
			if c := pendingChoiceOfKind(g, PendingChoiceTriggerPrompt); c != nil {
				t.Fatal("the \"you may\" was asked as the ability triggered, want at resolution (CR 603.5)")
			}
			g.WithWriteLock(func() { g.drainPendingTriggersAPNAPLocked() })
			item := optionalItemOnStack(g)
			if item == nil {
				t.Fatal("the optional trigger is not on the stack")
			}
			if item.Optional.Question != "Draw a card?" {
				t.Errorf("item question = %q, want the trigger's", item.Optional.Question)
			}

			c := passUntilTriggerPrompt(t, g)
			if c.OptionalItem != item.ID {
				t.Errorf("prompt names item %s, want the resolving item %s", c.OptionalItem, item.ID)
			}
			if c.Chooser != me.ID {
				t.Errorf("prompt chooser = %s, want the controller %s", c.Chooser, me.ID)
			}
			if !c.midResolution {
				t.Error("the prompt is not part of the resolution (#1289)")
			}
			if _, ok := g.StackMeta[item.ID]; ok {
				t.Error("the item is still waiting on the stack while it resolves")
			}
			if g.pausedResolutionTail == nil || g.pausedResolutionTail.item.ID != item.ID {
				t.Fatal("the resolution's tail is not parked behind the question")
			}
			if ran != 0 {
				t.Fatal("the effect ran before the answer")
			}
			// Nothing else happens while the question is open.
			if err := g.PassPriority(); !errors.Is(err, ErrChoicePending) {
				t.Errorf("PassPriority with the question open = %v, want ErrChoicePending", err)
			}

			if err := g.ResolveTriggerPrompt(c.ID, me.ID, tc.answer); err != nil {
				t.Fatalf("ResolveTriggerPrompt: %v", err)
			}
			if ran != tc.want {
				t.Errorf("effect ran %d times, want %d", ran, tc.want)
			}
			if g.pausedResolutionTail != nil || len(g.PendingChoices) != 0 {
				t.Errorf("the resolution did not finish: tail %v, choices %d", g.pausedResolutionTail, len(g.PendingChoices))
			}
			if !hasEventKind(g, EventResolutionFinished) {
				t.Error("no EventResolutionFinished: the log's \"resolved\" line never came")
			}
		})
	}
}

// TestOptionalTriggerCanBeRespondedToBeforeTheChoice: the trigger is on
// the stack, so every player gets priority over it before its
// controller decides (CR 603.5, 117.3).
func TestOptionalTriggerCanBeRespondedToBeforeTheChoice(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	ran := 0
	scheduleOptionalOnCast(g, me.ID, &ran)
	emitCast(g, me.ID, uuid.New())
	g.WithWriteLock(func() { g.drainPendingTriggersAPNAPLocked() })

	holder := g.Turn.PriorityHolder
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if pendingChoiceOfKind(g, PendingChoiceTriggerPrompt) != nil {
		t.Fatal("one pass asked the question; every player passes first")
	}
	if g.Turn.PriorityHolder == holder || optionalItemOnStack(g) == nil {
		t.Fatal("priority did not move on with the trigger still on the stack")
	}
	passUntilTriggerPrompt(t, g)
}

// TestOptionalTriggerPromptIsARestorePoint: the open question is plain
// data (the item rides the parked tail), so a restore into it still
// asks, and a "yes" there runs the effect.
func TestOptionalTriggerPromptIsARestorePoint(t *testing.T) {
	g := newRestorableGame(t)
	me := g.Seats[0]
	ran := 0
	scheduleOptionalOnCast(g, me.ID, &ran)
	emitCast(g, me.ID, uuid.New())
	c := passUntilTriggerPrompt(t, g)

	snap, restored := roundTrip(t, g)
	if !snap.Restorable() {
		t.Fatalf("the open \"you may\" is not a restore point: %+v", snap.Continuations)
	}
	rc := pendingChoiceOfKind(restored, PendingChoiceTriggerPrompt)
	if rc == nil || rc.ID != c.ID || rc.OptionalItem != c.OptionalItem {
		t.Fatalf("restored prompt = %+v, want %s naming %s", rc, c.ID, c.OptionalItem)
	}
	if err := restored.ResolveTriggerPrompt(rc.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveTriggerPrompt on the restored game: %v", err)
	}
	if ran != 1 {
		t.Errorf("effect ran %d times after the restored yes, want 1", ran)
	}
	if restored.pausedResolutionTail != nil {
		t.Error("the restored resolution did not finish")
	}
}

// TestOptionalTargetedTriggerPicksItsTargetBeforeTheStack: CR 603.3d
// still chooses targets as the ability is put on the stack, and only
// the "you may" waits for resolution.
func TestOptionalTargetedTriggerPicksItsTargetBeforeTheStack(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	source := Card{InstanceID: uuid.New(), Name: "Probe Sage", Controller: me.ID, Owner: me.ID}
	g.Battlefield.PushTop(source)
	victim := Card{InstanceID: uuid.New(), Name: "Probe Rock", TypeLine: "Artifact", Controller: opp.ID, Owner: opp.ID}
	g.Battlefield.PushTop(victim)
	var hit []TargetRef
	ability := TriggeredAbility{
		Key:            "Probe Sage — destroy target artifact",
		OptionalPrompt: &TriggerOptionalPrompt{Question: "Destroy it?"},
		Targets:        probeArtifactTarget(),
		Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			return newTriggeredItemForTest(source, "Probe Sage — destroy target artifact", func(_ *Game, it *StackItem) error {
				hit = append(hit, it.Targets...)
				return nil
			})
		},
	}
	g.WithWriteLock(func() {
		g.dispatchTriggerLocked(Event{Kind: EventETB, CardID: source.InstanceID}, source, source.Effective(), ability)
	})
	pick := pendingChoiceOfKind(g, PendingChoicePickTarget)
	if pick == nil {
		t.Fatal("no CR 603.3d target pick as the trigger went on the stack")
	}
	if pendingChoiceOfKind(g, PendingChoiceTriggerPrompt) != nil {
		t.Fatal("the \"you may\" was asked before the target pick")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, TargetRef{Kind: TargetCard, ID: victim.InstanceID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	c := passUntilTriggerPrompt(t, g)
	if err := g.ResolveTriggerPrompt(c.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveTriggerPrompt: %v", err)
	}
	if len(hit) != 1 || hit[0].ID != victim.InstanceID {
		t.Errorf("effect saw targets %+v, want the picked artifact", hit)
	}
}

// TestOptionalTriggerWithNoLegalTargetIsNotPutOnTheStack: a "you may"
// whose required target clause has nothing to point at is removed as it
// would be put on the stack (CR 603.3d), with no question at all.
func TestOptionalTriggerWithNoLegalTargetIsNotPutOnTheStack(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	source := Card{InstanceID: uuid.New(), Name: "Probe Sage", Controller: me.ID, Owner: me.ID}
	g.Battlefield.PushTop(source)
	ability := TriggeredAbility{
		Key:            "Probe Sage — destroy target artifact",
		OptionalPrompt: &TriggerOptionalPrompt{Question: "Destroy it?"},
		Targets:        probeArtifactTarget(),
		Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			t.Fatal("Build ran for a trigger with no legal target")
			return nil
		},
	}
	g.WithWriteLock(func() {
		g.dispatchTriggerLocked(Event{Kind: EventETB, CardID: source.InstanceID}, source, source.Effective(), ability)
	})
	if len(g.PendingChoices) != 0 || len(g.PendingTriggers) != 0 {
		t.Errorf("a targetless optional trigger left %d prompts and %d items", len(g.PendingChoices), len(g.PendingTriggers))
	}
}

// probeArtifactTarget is "target artifact".
func probeArtifactTarget() *TargetSpec {
	return &TargetSpec{
		Mode: "artifact", Label: "target artifact", Zones: []ZoneKind{ZoneBattlefield},
		CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool { return c.IsArtifact() },
		Min:    1, Max: 1,
	}
}

func hasEventKind(g *Game, kind EventKind) bool {
	for i := range g.Events {
		if g.Events[i].Kind == kind {
			return true
		}
	}
	return false
}
