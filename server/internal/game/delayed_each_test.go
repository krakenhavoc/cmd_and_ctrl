package game

import (
	"encoding/json"
	"fmt"
	"slices"
	"testing"

	"github.com/google/uuid"
)

// delayed_each_test.go — ADR 0187 amendment 2026-10-16 (#3531): a
// delayed trigger that triggers once for each object its event concerns
// (an each-condition, CR 603.2c), and one whose body announces its own
// target as each instance goes on the stack (CR 603.3d).

// testEachCondition registers a one-off each-condition.
func testEachCondition(fn func(ev Event, dt *DelayedTrigger, g *Game) []uuid.UUID) ConditionRef {
	key := fmt.Sprintf("%sper-object-%d", testEffectKeyPrefix, testEffectKeySeq.Add(1))
	return DelayedEachCondition(key, func(ev Event, dt *DelayedTrigger, g *Game, _ EffectParams) []uuid.UUID {
		return fn(ev, dt, g)
	})
}

// eachProbe schedules a delayed trigger on the controller's casts whose
// each-condition answers `objects`, and records each resolution's
// payload.
func eachProbe(g *Game, controller uuid.UUID, repeats bool, objects []uuid.UUID, seen *[]uuid.UUID) {
	g.WithWriteLock(func() {
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller:   controller,
			SourceCardID: uuid.New(),
			Label:        "probe — once for each object",
			On:           []EventKind{EventCast},
			Repeats:      repeats,
			Condition: testEachCondition(func(ev Event, dt *DelayedTrigger, _ *Game) []uuid.UUID {
				if ev.Actor != dt.Controller {
					return nil
				}
				return objects
			}),
			Body: testBody(func(_ *Game, item *StackItem) error {
				for _, p := range item.Payload {
					*seen = append(*seen, p.ID)
				}
				return nil
			}),
		})
	})
}

// One event that concerns two objects triggers twice, each instance with
// its own object as "it", and a repeating trigger does it again on the
// next event.
func TestEachConditionTriggersOnceForEachObject(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	a, b := uuid.New(), uuid.New()
	var seen []uuid.UUID
	eachProbe(g, me, true, []uuid.UUID{a, b}, &seen)

	emitCast(g, me, uuid.New())
	settleStack(t, g)
	if len(seen) != 2 || !slices.Contains(seen, a) || !slices.Contains(seen, b) {
		t.Fatalf("payloads %v after one event, want one instance for each of %v and %v", seen, a, b)
	}
	emitCast(g, me, uuid.New())
	settleStack(t, g)
	if len(seen) != 4 {
		t.Errorf("%d instances after two events, want 4 (CR 603.7b: this turn)", len(seen))
	}
	// An event the condition answers nothing for is not its event.
	emitCast(g, g.Seats[1].ID, uuid.New())
	settleStack(t, g)
	if len(seen) != 4 {
		t.Errorf("a non-matching event fired it: %d instances", len(seen))
	}
}

// A one-shot trigger (no stated duration) triggers once, then ceases to
// exist (CR 603.7b).
func TestOneShotEachConditionTriggersOnce(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	a, b := uuid.New(), uuid.New()
	var seen []uuid.UUID
	eachProbe(g, me, false, []uuid.UUID{a, b}, &seen)

	emitCast(g, me, uuid.New())
	settleStack(t, g)
	emitCast(g, me, uuid.New())
	settleStack(t, g)
	if len(seen) != 1 || seen[0] != a {
		t.Fatalf("payloads %v, want one instance for %v", seen, a)
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("%d delayed triggers still queued, want 0", len(g.DelayedTriggers))
	}
}

// targetedProbe schedules a repeating trigger on the controller's casts
// whose body announces "target creature", and records the target each
// resolution saw.
func targetedProbe(g *Game, controller uuid.UUID, saw *[]uuid.UUID) {
	g.WithWriteLock(func() {
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller:   controller,
			SourceCardID: uuid.New(),
			Label:        "probe — target creature",
			On:           []EventKind{EventCast},
			Repeats:      true,
			Condition: testCondition(func(ev Event, dt *DelayedTrigger, _ *Game) bool {
				return ev.Actor == dt.Controller
			}),
			Body: testTargetedDelayedBody(func(_ *Game, item *StackItem) error {
				for _, t := range item.Targets {
					*saw = append(*saw, t.ID)
				}
				return nil
			}),
		})
	})
}

func testTargetedDelayedBody(fn func(g *Game, item *StackItem) error) BodyRef {
	key := fmt.Sprintf("%stargeted-%d", testEffectKeyPrefix, testEffectKeySeq.Add(1))
	return TargetedDelayedBody(key, func(g *Game, item *StackItem, _ EffectParams) error { return fn(g, item) },
		func(uuid.UUID, EffectParams) *TargetSpec { return creatureSpec() })
}

// The target is chosen as the instance goes on the stack, by the
// trigger's controller, from the board as it is then (CR 603.3d).
func TestTargetedDelayedBodyAsksForItsTarget(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	bear := pushCreatureFor(t, g, g.Seats[1], 2, 2)
	var saw []uuid.UUID
	targetedProbe(g, me, &saw)

	emitCast(g, me, uuid.New())
	prompt := pickTargetPrompt(g)
	if prompt == nil {
		t.Fatal("no pick_target prompt when the trigger fired")
	}
	if prompt.Chooser != me || !slices.Contains(prompt.PickTargetCards, bear) {
		t.Fatalf("prompt chooser %v options %v, want %v choosing among %v", prompt.Chooser, prompt.PickTargetCards, me, bear)
	}
	if err := g.ResolvePickTarget(prompt.ID, me, TargetRef{Kind: TargetCard, ID: bear}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	settleStack(t, g)
	if len(saw) != 1 || saw[0] != bear {
		t.Fatalf("the body saw targets %v, want %v", saw, bear)
	}
}

// With no legal target the instance is removed and nothing is asked
// (CR 603.3d).
func TestTargetedDelayedBodyWithNoTargetIsRemoved(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	var saw []uuid.UUID
	targetedProbe(g, me, &saw)

	emitCast(g, me, uuid.New())
	if p := pickTargetPrompt(g); p != nil {
		t.Fatalf("a target prompt was opened with no creature: %+v", p)
	}
	settleStack(t, g)
	if len(saw) != 0 {
		t.Errorf("the body ran %d times with no target", len(saw))
	}
}

// An instance waiting on the stack with its target chosen is a restore
// point, and the restored item still re-checks its target.
func TestTargetedDelayedItemSurvivesARestore(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	bear := pushCreatureFor(t, g, g.Seats[1], 2, 2)
	var saw []uuid.UUID
	targetedProbe(g, me, &saw)

	emitCast(g, me, uuid.New())
	prompt := pickTargetPrompt(g)
	if prompt == nil {
		t.Fatal("no pick_target prompt")
	}
	if err := g.ResolvePickTarget(prompt.ID, me, TargetRef{Kind: TargetCard, ID: bear}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("a table with a targeted delayed trigger on the stack is not a restore point: %+v", snap.Continuations)
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	var back GameSnapshot
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatal(err)
	}
	restored, err := back.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	var spec *TargetSpec
	for _, it := range restored.StackMeta {
		if it != nil && it.targetSpec != nil {
			spec = it.targetSpec
		}
	}
	if spec == nil {
		t.Fatal("the restored item has no target clause to re-check (CR 608.2b)")
	}
	settleStack(t, restored)
	if len(saw) != 1 || saw[0] != bear {
		t.Errorf("the restored item saw targets %v, want %v", saw, bear)
	}
}

// A targeted body cannot also carry Cards: its announced targets would
// overwrite them.
func TestTargetedDelayedBodyRefusesCards(t *testing.T) {
	g := newActiveGame(t)
	defer func() {
		if recover() == nil {
			t.Error("scheduling a targeted body with Cards did not fault")
		}
	}()
	g.WithWriteLock(func() {
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller: g.Seats[0].ID,
			Label:      "probe",
			On:         []EventKind{EventCast},
			Cards:      []uuid.UUID{uuid.New()},
			Body:       testTargetedDelayedBody(func(*Game, *StackItem) error { return nil }),
		})
	})
}
