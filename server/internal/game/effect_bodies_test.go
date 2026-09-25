package game

import (
	"encoding/json"
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
func testBody(fn func(g *Game, item *StackItem) error) BodyRef {
	key := fmt.Sprintf("%sbody-%d", testEffectKeyPrefix, testEffectKeySeq.Add(1))
	return SimpleDelayedBody(key, fn)
}

// testCondition registers a one-off test event condition.
func testCondition(fn func(ev Event, dt *DelayedTrigger, g *Game) bool) ConditionRef {
	key := fmt.Sprintf("%scondition-%d", testEffectKeyPrefix, testEffectKeySeq.Add(1))
	return DelayedCondition(key, func(ev Event, dt *DelayedTrigger, g *Game, _ EffectParams) bool {
		return fn(ev, dt, g)
	})
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
		SimpleDelayedBody(key.Key(), func(*Game, *StackItem) error { return nil })
	}()
}

func TestAnAliasResolvesTheOldKey(t *testing.T) {
	newKey := testBody(func(*Game, *StackItem) error { return nil })
	oldKey := fmt.Sprintf("%srenamed-%d", testEffectKeyPrefix, testEffectKeySeq.Add(1))
	EffectAlias(oldKey, newKey.Key())
	if !KnownEffectBody(oldKey) {
		t.Errorf("the alias %q does not resolve to %q", oldKey, newKey.Key())
	}
}

// TestAForgottenBodyFailsATestAndDropsInProduction: Body is a typed
// BodyRef, so an unregistered key cannot be written at all; what is
// left is the zero ref — a forgotten `Body:` that still compiles. In a
// test binary effectKeyFault panics, so the first test that schedules
// one fails; production logs and drops (#1568 review). The panic IS
// the test-binary branch, which is the only one a test can run.
func TestAForgottenBodyFailsATestAndDropsInProduction(t *testing.T) {
	g := newActiveGame(t)
	defer func() {
		if recover() == nil {
			t.Error("a delayed trigger with no body was queued without a fault")
		}
		if n := len(g.DelayedTriggers); n != 0 {
			t.Errorf("%d delayed triggers queued; a bodiless one must be dropped", n)
		}
	}()
	g.WithWriteLock(func() {
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{Controller: g.Seats[0].ID, At: StepEnd, Label: "x"})
	})
}

// TestAnInvalidSpellFilterIsRefused: CastFilter.Types is CR 205.2a's
// closed list, spelled exactly (#1568 review).
func TestAnInvalidSpellFilterIsRefused(t *testing.T) {
	for _, bad := range [][]string{{""}, {"art"}, {"instant"}, {"Instnat"}, {"Legendary"}} {
		f := CastFilter{Types: bad}
		if f.Valid() {
			t.Errorf("CastFilter%v is Valid", bad)
		}
		if f.Matches(Card{TypeLine: "Artifact Creature — Golem"}) || f.Matches(Card{TypeLine: "Instant"}) {
			t.Errorf("CastFilter%v matched a card", bad)
		}
	}
	g := newActiveGame(t)
	func() {
		defer func() {
			if recover() == nil {
				t.Error("a delayed trigger with an invalid filter was queued without a fault")
			}
		}()
		g.WithWriteLock(func() {
			g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
				Controller: g.Seats[0].ID, On: []EventKind{EventCast}, Label: "x",
				Body:       testBody(func(*Game, *StackItem) error { return nil }),
				CondParams: EffectParams{Filter: CastFilter{Types: []string{"art"}}},
			})
		})
	}()
	// And at restore.
	good := newActiveGame(t)
	good.WithWriteLock(func() {
		good.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller: good.Seats[0].ID, On: []EventKind{EventCast}, Label: "x",
			Body:       testBody(func(*Game, *StackItem) error { return nil }),
			CondParams: EffectParams{Filter: CastFilter{Types: []string{"Instant"}}},
		})
	})
	snap := good.CaptureSnapshot()
	snap.DelayedTriggers[0].CondParams.Filter.Types = []string{"art"}
	if _, err := snap.RestoreStrict(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Errorf("RestoreStrict of an invalid filter: err = %v, want ErrUnknownEffectKey", err)
	}
}

// TestCastFilterMatchesWholeTypes: exact type names, not substrings.
func TestCastFilterMatchesWholeTypes(t *testing.T) {
	f := CastFilter{Types: []string{"Instant", "Sorcery"}}
	for _, tc := range []struct {
		typeLine string
		want     bool
	}{
		{"Instant", true},
		{"Sorcery", true},
		{"Kindred Instant — Rogue", true},
		{"Legendary Sorcery", true},
		{"Creature — Elf", false},
		{"Artifact", false},
		{"Enchantment — Aura", false},
	} {
		if got := f.Matches(Card{TypeLine: tc.typeLine}); got != tc.want {
			t.Errorf("Matches(%q) = %v, want %v", tc.typeLine, got, tc.want)
		}
	}
	art := CastFilter{Types: []string{"Artifact"}}
	if !art.Matches(Card{TypeLine: "Artifact Creature — Golem"}) {
		t.Error("an Artifact filter does not match an artifact creature")
	}
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

// TestUnknownParamsFieldsAreRefused is the #1568 review's first
// finding, tier 1's lesson again: encoding/json drops a key it has no
// field for, and inside `params`, `condParams`, a `filter` or an
// `object` an unknown key is exactly what a newer body's data looks
// like. Refused, not dropped — on delayed triggers and on stack items.
func TestUnknownParamsFieldsAreRefused(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	body := testBody(func(*Game, *StackItem) error { return nil })
	g.WithWriteLock(func() {
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller: me, At: StepEnd, Label: "params", Body: body,
			Params: EffectParams{Amount: 3, Object: ObjectRef{ID: me, Epoch: 1}},
		})
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller: me, On: []EventKind{EventCast}, Label: "condParams", Body: body,
			CondParams: EffectParams{Filter: CastFilter{Types: []string{"Instant"}}},
		})
		id := uuid.New()
		if g.StackMeta == nil {
			g.StackMeta = map[uuid.UUID]*StackItem{}
		}
		g.StackMeta[id] = &StackItem{
			ID: id, Kind: StackItemTriggered, Controller: me, Label: "keyed item",
			Body: body.Key(), Params: EffectParams{Amount: 5},
			Effect: bodyEffect(body.Key(), EffectParams{Amount: 5}),
		}
	})
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	type corruption func(snap map[string]any)
	dt := func(snap map[string]any, i int) map[string]any {
		return snap["delayedTriggers"].([]any)[i].(map[string]any)
	}
	cases := map[string]corruption{
		"params": func(s map[string]any) { dt(s, 0)["params"].(map[string]any)["futureCount"] = 7 },
		"params.object": func(s map[string]any) {
			dt(s, 0)["params"].(map[string]any)["object"].(map[string]any)["zone"] = "exile"
		},
		"condParams": func(s map[string]any) { dt(s, 1)["condParams"].(map[string]any)["futureField"] = 1 },
		"condParams.filter": func(s map[string]any) {
			dt(s, 1)["condParams"].(map[string]any)["filter"].(map[string]any)["supertypes"] = []string{"Legendary"}
		},
		"stack params": func(s map[string]any) {
			s["stackMeta"].([]any)[0].(map[string]any)["params"].(map[string]any)["futureCount"] = 7
		},
	}
	for name, corrupt := range cases {
		t.Run(name, func(t *testing.T) {
			var generic map[string]any
			if err := json.Unmarshal(raw, &generic); err != nil {
				t.Fatal(err)
			}
			corrupt(generic)
			bad, err := json.Marshal(generic)
			if err != nil {
				t.Fatal(err)
			}
			var snap GameSnapshot
			if err := json.Unmarshal(bad, &snap); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if _, err := snap.RestoreStrict(); !errors.Is(err, ErrUnknownEffectKey) {
				t.Errorf("RestoreStrict: err = %v, want ErrUnknownEffectKey", err)
			}
		})
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	if _, err := snap.RestoreStrict(); err != nil {
		t.Errorf("the intact file was refused: %v", err)
	}
}

// TestACopiedKeyedAbilityIsStillData: a Strionic Resonator copy of a
// fired delayed trigger (a Mana Drain refund waiting on the stack) is
// keyed too, so the table stays a restore point (#1568 review).
func TestACopiedKeyedAbilityIsStillData(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	body := testBody(func(*Game, *StackItem) error { return nil })
	id := uuid.New()
	g.WithWriteLock(func() {
		if g.StackMeta == nil {
			g.StackMeta = map[uuid.UUID]*StackItem{}
		}
		g.StackMeta[id] = &StackItem{
			ID: id, Kind: StackItemTriggered, Controller: me, Owner: me, Label: "refund",
			Body: body.Key(), Params: EffectParams{Amount: 4},
			Effect: bodyEffect(body.Key(), EffectParams{Amount: 4}),
		}
		if err := g.CopyAbilityForEffect(id, me, false); err != nil {
			t.Fatalf("CopyAbilityForEffect: %v", err)
		}
	})
	var copied *StackItem
	for k, it := range g.StackMeta {
		if k != id {
			copied = it
		}
	}
	if copied == nil {
		t.Fatal("no copy was made")
	}
	if copied.Body != body.Key() || copied.Params.Amount != 4 {
		t.Errorf("the copy's Body %q / Params %+v, want the original's", copied.Body, copied.Params)
	}
	if snap := g.CaptureSnapshot(); !snap.Restorable() {
		t.Errorf("a copied keyed ability blocks the restore point: %+v", snap.Continuations)
	}
}

// TestALedgerLineResolvesOnlyAsItsOwnKind: a "body X" line is not kept
// alive by a condition that happens to be named X (#1568 review).
func TestALedgerLineResolvesOnlyAsItsOwnKind(t *testing.T) {
	cond := testCondition(func(Event, *DelayedTrigger, *Game) bool { return false })
	body := testBody(func(*Game, *StackItem) error { return nil })
	if ledgerLineResolves("body " + cond.Key()) {
		t.Error("a body line resolved through a condition of the same name")
	}
	if ledgerLineResolves("condition " + body.Key()) {
		t.Error("a condition line resolved through a body of the same name")
	}
	if !ledgerLineResolves("body "+body.Key()) || !ledgerLineResolves("condition "+cond.Key()) {
		t.Error("a registered key's own line does not resolve")
	}
}
