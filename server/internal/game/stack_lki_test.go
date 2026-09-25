package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// stack_lki_test.go — #1255. The last-known record of a spell that
// left the stack without resolving, tested as a shape: taken when a
// spell is countered, never widening the strict copy lookup, cleared
// at the turn boundary, rewound by an undo, and not carried by a
// snapshot. The cards that read it are pinned in
// cards/effects/storm_test.go and lki_copy_test.go.

// pushLKISpell puts a plain instant on the stack with its meta, the
// way a cast leaves it, and returns its instance ID.
func pushLKISpell(t *testing.T, g *Game, controller uuid.UUID, name string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		g.Stack.PushTop(Card{
			InstanceID: id, Name: name, TypeLine: "Instant",
			Owner: controller, Controller: controller,
		})
		if g.StackMeta == nil {
			g.StackMeta = make(map[uuid.UUID]*StackItem)
		}
		g.StackMeta[id] = &StackItem{
			ID: id, Kind: StackItemSpell, Controller: controller, Owner: controller,
			SourceCardID: id, XValue: 4, Seq: g.nextStackSeqLocked(),
		}
	})
	return id
}

func counterForTest(t *testing.T, g *Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.CounterTargetForEffect(id); err != nil {
			t.Fatalf("CounterTargetForEffect: %v", err)
		}
	})
}

// TestACounteredSpellIsRememberedAsItStoodOnTheStack — the record holds
// the card and the announce-time choices (X here), and the strict
// lookup is exactly as strict as it was.
func TestACounteredSpellIsRememberedAsItStoodOnTheStack(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushLKISpell(t, g, me.ID, "Fireball")
	counterForTest(t, g, id)

	g.WithWriteLock(func() {
		card, item, ok := g.lastKnownSpellLocked(id)
		if !ok {
			t.Fatal("the countered spell left no record")
		}
		if card.Name != "Fireball" || item.XValue != 4 || item.Kind != StackItemSpell {
			t.Errorf("record = %q / X=%d / %s, want the spell as it stood", card.Name, item.XValue, item.Kind)
		}
		// The item handed out is a copy: writing to it cannot reach
		// the record.
		item.XValue = 99
		if _, again, _ := g.lastKnownSpellLocked(id); again.XValue != 4 {
			t.Errorf("the record was mutated through the returned item: X=%d", again.XValue)
		}

		if err := g.CopySpellForEffect(id, me.ID, false, nil); err != ErrCardNotFound {
			t.Errorf("strict CopySpellForEffect on a countered spell = %v, want ErrCardNotFound (CR 608.2b)", err)
		}
		if g.Stack.Size() != 0 {
			t.Fatalf("the strict lookup made a copy")
		}
		if err := g.CopyLastKnownSpellForEffect(id, me.ID, false, nil); err != nil {
			t.Errorf("CopyLastKnownSpellForEffect: %v", err)
		}
		if g.Stack.Size() != 1 {
			t.Fatalf("the last-known copy put %d objects on the stack, want 1", g.Stack.Size())
		}
		cp := g.Stack.Cards[0]
		meta := g.StackMeta[cp.InstanceID]
		if cp.InstanceID == id || cp.Name != "Fireball" || meta == nil || !meta.IsCopy || meta.XValue != 4 {
			t.Errorf("copy = %q (same id: %v), meta %+v; want a new Fireball copy with X=4", cp.Name, cp.InstanceID == id, meta)
		}
	})
}

// TestAResolvedSpellLeavesNoRecord — resolution is not a way out the
// record covers: Game.resolving (#920) answers for "copy this spell",
// and nothing else can name a resolved spell.
func TestAResolvedSpellLeavesNoRecord(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushLKISpell(t, g, me.ID, "Opt")
	g.WithWriteLock(func() {
		if err := g.resolveTopOfStackLocked(); err != nil {
			t.Fatalf("resolveTopOfStackLocked: %v", err)
		}
		if _, _, ok := g.lastKnownSpellLocked(id); ok {
			t.Error("a resolved spell was recorded")
		}
	})
}

// TestTheRecordIsClearedAtTheTurnBoundary — nothing can name a spell
// from a previous turn: the stack is empty when a turn ends.
func TestTheRecordIsClearedAtTheTurnBoundary(t *testing.T) {
	g := newActiveGame(t)
	id := pushLKISpell(t, g, g.Seats[0].ID, "Opt")
	counterForTest(t, g, id)
	g.WithWriteLock(func() {
		g.onTurnBeganLocked()
		if _, _, ok := g.lastKnownSpellLocked(id); ok {
			t.Error("the record outlived its turn")
		}
	})
}

// TestTheRecordRewindsWithAnUndo — an undo taken before the counter
// has no record; one taken after has it, and later inserts on the live
// game do not reach the stored snapshot.
func TestTheRecordRewindsWithAnUndo(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	first := pushLKISpell(t, g, me.ID, "Opt")
	second := pushLKISpell(t, g, me.ID, "Shock")
	before := g.Clone()

	counterForTest(t, g, second)
	afterOne := g.Clone()
	counterForTest(t, g, first)

	g.WithWriteLock(func() { g.RestoreFrom(afterOne) })
	g.WithWriteLock(func() {
		if _, _, ok := g.lastKnownSpellLocked(second); !ok {
			t.Error("the undo point after the counter lost the record")
		}
		if _, _, ok := g.lastKnownSpellLocked(first); ok {
			t.Error("a counter made after the undo point leaked into it")
		}
	})
	// Countering again on the restored game must not write through to
	// the snapshot it was restored from.
	counterForTest(t, g, first)
	afterOne.mu.RLock()
	_, leaked := afterOne.lastKnownStack[first]
	afterOne.mu.RUnlock()
	if leaked {
		t.Error("the restored game shares its record with the undo snapshot")
	}

	g.WithWriteLock(func() { g.RestoreFrom(before) })
	g.WithWriteLock(func() {
		if len(g.lastKnownStack) != 0 {
			t.Errorf("the undo point before any counter has %d records", len(g.lastKnownStack))
		}
	})
}

// TestTheRecordIsSnapshotted — `carried` in snapshot_drift_test.go
// since ADR 0041 phase 3 tier 4 (#1497, P9): the record's readers become
// data, so a restore point taken after a counterspell carries the
// countered spell's last-known information, and a copy effect that names
// it still finds it after a deploy.
func TestTheRecordIsSnapshotted(t *testing.T) {
	g := newActiveGame(t)
	id := pushLKISpell(t, g, g.Seats[0].ID, "Opt")
	counterForTest(t, g, id)
	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("a countered spell's record blocks the restore point: %+v", snap.Continuations)
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	var decoded GameSnapshot
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	restored, err := decoded.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	restored.WithWriteLock(func() {
		card, item, ok := restored.lastKnownSpellLocked(id)
		if !ok {
			t.Fatal("the restored game has no record of the countered spell")
		}
		if card.Name != "Opt" || item.XValue != 4 || item.Kind != StackItemSpell {
			t.Errorf("record came back as %q, X=%d, kind %s", card.Name, item.XValue, item.Kind)
		}
		if err := restored.CopyLastKnownSpellForEffect(id, restored.Seats[0].ID, false, nil); err != nil {
			t.Errorf("CopyLastKnownSpellForEffect after the restore: %v", err)
		}
	})
}
