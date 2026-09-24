package game

import (
	"errors"
	"flag"
	"fmt"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
)

// effect_bodies_test.go pins ADR 0041 phase 3's tier 2 (#1497): the
// registry of delayed-trigger bodies and conditions, and the
// append-only ledger of their keys.

// testEffectKeyPrefix is the namespace test-only bodies register under.
// The ledger ignores it: a test body is never in a real restore point.
const testEffectKeyPrefix = effectKeyTestPrefix

var testEffectKeySeq atomic.Int64

// testBody registers a one-off test body and returns its key, so a test
// can still schedule a delayed trigger whose effect is a closure over
// its own counters — the production path cannot.
func testBody(fn func(g *Game, item *StackItem) error) string {
	key := fmt.Sprintf("%sbody-%d", testEffectKeyPrefix, testEffectKeySeq.Add(1))
	SimpleDelayedBody(key, fn)
	return key
}

// testCondition registers a one-off test event condition.
func testCondition(fn func(ev Event, dt *DelayedTrigger, g *Game) bool) string {
	key := fmt.Sprintf("%scondition-%d", testEffectKeyPrefix, testEffectKeySeq.Add(1))
	DelayedCondition(key, func(ev Event, dt *DelayedTrigger, g *Game, _ EffectParams) bool {
		return fn(ev, dt, g)
	})
	return key
}

var updateEffectKeys = flag.Bool("update-effect-keys", false,
	"append newly registered effect keys to testdata/effect_keys.txt (never removes a line)")

// TestEveryPersistedEffectKeyResolves is the ledger's engine half: every
// body and condition the ENGINE registers is in testdata/effect_keys.txt.
// The catalog package runs the full check — which also fails on a
// ledger line that no longer resolves — because only it has every
// production body registered. See CheckEffectKeyLedger.
func TestEveryPersistedEffectKeyResolves(t *testing.T) {
	missing, _, err := CheckEffectKeyLedger(filepath.Join("testdata", "effect_keys.txt"), *updateEffectKeys, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range missing {
		t.Errorf("%q is registered but not in testdata/effect_keys.txt — append it with "+
			"go test ./internal/cards/effects -run TestEveryPersistedEffectKeyResolves -args -update-effect-keys", line)
	}
}

func TestEffectKeysAreShapedAndUnique(t *testing.T) {
	for _, bad := range []string{"NoSlash", "a/B", "a//b", "/b", "a/", "batch 1/x"} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("key %q registered without a panic", bad)
				}
			}()
			SimpleDelayedBody(bad, func(*Game, *StackItem) error { return nil })
		}()
	}
	key := testBody(func(*Game, *StackItem) error { return nil })
	func() {
		defer func() {
			if recover() == nil {
				t.Error("a duplicate key registered without a panic")
			}
		}()
		SimpleDelayedBody(key, func(*Game, *StackItem) error { return nil })
	}()
}

func TestAnAliasResolvesTheOldKey(t *testing.T) {
	newKey := testBody(func(*Game, *StackItem) error { return nil })
	oldKey := fmt.Sprintf("%srenamed-%d", testEffectKeyPrefix, testEffectKeySeq.Add(1))
	EffectAlias(oldKey, newKey)
	if !KnownEffectBody(oldKey) {
		t.Errorf("the alias %q does not resolve to %q", oldKey, newKey)
	}
}

func TestSchedulingAnUnregisteredBodyPanics(t *testing.T) {
	g := newActiveGame(t)
	defer func() {
		if recover() == nil {
			t.Error("a delayed trigger naming an unregistered body was queued")
		}
	}()
	g.WithWriteLock(func() {
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller: g.Seats[0].ID, At: StepEnd, Label: "x", Body: "nobody/registered-this",
		})
	})
}

// TestADelayedTriggerIsARestorePoint is tier 2's point: a table with a
// delayed trigger waiting — before it fires and after it is on the
// stack — is a full restore point, and the restored trigger still does
// its work through the running binary.
func TestADelayedTriggerIsARestorePoint(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	var fired atomic.Int64
	body := testBody(func(*Game, *StackItem) error { fired.Add(1); return nil })
	g.WithWriteLock(func() {
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller: me, At: StepEnd, Label: "test — at the next end step", Body: body,
			Params: EffectParams{Amount: 3, Cost: "{2}"},
		})
	})
	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("a queued delayed trigger blocks the restore point: %+v", snap.Continuations)
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	if len(restored.DelayedTriggers) != 1 || restored.DelayedTriggers[0].Params.Amount != 3 {
		t.Fatalf("the restored queue is %+v", restored.DelayedTriggers)
	}
	advanceTo(t, restored, StepEnd)
	if len(restored.StackMeta)+len(restored.PendingTriggers) == 0 {
		t.Fatal("the restored trigger did not fire at the end step")
	}
	// It fired: it is a keyed item now, and still a restore point.
	stack := restored.CaptureSnapshot()
	if !stack.Restorable() {
		t.Fatalf("a fired delayed trigger on the stack blocks the restore point: %+v", stack.Continuations)
	}
	again, err := stack.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict after firing: %v", err)
	}
	settleStack(t, again)
	if fired.Load() != 1 {
		t.Errorf("the restored fired trigger resolved %d times, want 1", fired.Load())
	}
}

func TestAStackItemWithAnUnknownBodyIsRefused(t *testing.T) {
	g := newActiveGame(t)
	snap := g.CaptureSnapshot()
	snap.PendingTriggers = append(snap.PendingTriggers, stackItemSnapshot{ID: uuid.New(), Body: "nobody/registered-this"})
	if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Errorf("err = %v, want ErrUnknownEffectKey", err)
	}
}
