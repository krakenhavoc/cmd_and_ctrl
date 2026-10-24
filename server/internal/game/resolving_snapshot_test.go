package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// resolving_snapshot_test.go — the resolving slot across a restore
// (#4270). A keyed prompt carries no resume frame, so a table waiting on
// one is a restore point, and its answer reads Game.resolving.

func resolvingSpellSlot(g *Game) (*resolvingItem, uuid.UUID) {
	me, other := g.Seats[0].ID, g.Seats[1].ID
	id := uuid.New()
	return &resolvingItem{
		item: &StackItem{
			ID: id, Kind: StackItemSpell, Controller: me, Owner: me,
			SourceCardID: id, Label: "Test Bolt", XValue: 3, Seq: 77,
			Targets: []TargetRef{{Kind: TargetPlayer, ID: other}},
		},
		card: Card{InstanceID: id, Name: "Test Bolt", OracleID: "oracle-test-bolt",
			TypeLine: "Instant", Owner: me, Controller: me},
		hasCard:     true,
		sourceEpoch: -1,
	}, id
}

func TestResolvingSpellSurvivesARestore(t *testing.T) {
	g := newRestorableGame(t)
	slot, id := resolvingSpellSlot(g)
	g.WithWriteLock(func() { g.resolving = slot })

	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("a spent resolving slot made the game a non-restore point: %+v", snap.Continuations)
	}
	_, restored := roundTrip(t, g)

	var card Card
	var item *StackItem
	var ok bool
	restored.WithWriteLock(func() { card, item, ok = restored.resolvingSpellLocked(id) })
	if !ok {
		t.Fatal("the restored game lost the resolving spell")
	}
	if card.Name != "Test Bolt" || card.InstanceID != id {
		t.Errorf("restored card = %q %v, want Test Bolt %v", card.Name, card.InstanceID, id)
	}
	if item.Controller != g.Seats[0].ID || item.XValue != 3 || len(item.Targets) != 1 ||
		item.Targets[0].ID != g.Seats[1].ID || item.Label != "Test Bolt" {
		t.Errorf("restored item = %+v, want the controller, X, target and label it had", item)
	}
	// Every other lookup stays narrow: another ID does not reach the slot.
	restored.WithWriteLock(func() {
		if _, _, ok := restored.resolvingSpellLocked(uuid.New()); ok {
			t.Error("the restored slot answered for an ID that is not its own")
		}
	})
}

func TestResolvingAbilityKeepsItsSourceEpochAcrossARestore(t *testing.T) {
	g := newRestorableGame(t)
	me := g.Seats[0].ID
	src := uuid.New()
	item := &StackItem{
		ID: uuid.New(), Kind: StackItemTriggered, Controller: me, Owner: me,
		SourceCardID: src, SourceObject: ObjectRef{ID: src, Epoch: 4}, Label: "Test trigger", Seq: 78,
	}
	g.WithWriteLock(func() { g.resolving = &resolvingItem{item: item, sourceEpoch: 4} })

	_, restored := roundTrip(t, g)
	restored.WithWriteLock(func() {
		got, ok := restored.stackAbilityLocked(item.ID)
		if !ok || got.Label != "Test trigger" || got.Controller != me {
			t.Fatalf("the restored slot gave %+v, %v for the resolving ability", got, ok)
		}
		start, ok := restored.resolvingSourceEpochLocked(got)
		if !ok || start != 4 {
			t.Errorf("resolvingSourceEpochLocked = %d, %v, want 4, true", start, ok)
		}
		if restored.resolving.hasCard {
			t.Error("an ability item came back holding a card")
		}
		if cause := restored.resolutionCauseLocked(); cause.Kind != MoveCauseEffect ||
			cause.Controller != me || cause.Item != item.ID {
			t.Errorf("resolutionCauseLocked = %+v, want the ability's effect cause", cause)
		}
	})
}

// A restored slot is cleared at the same point a live one is: the next
// event batch. The slot is held for one batch, and the batch counter
// restores with it.
func TestRestoredResolvingSlotIsClearedByTheNextEventBatch(t *testing.T) {
	g := newRestorableGame(t)
	slot, id := resolvingSpellSlot(g)
	g.WithWriteLock(func() { g.resolving = slot })

	_, restored := roundTrip(t, g)
	restored.WithWriteLock(func() {
		if _, _, ok := restored.resolvingSpellLocked(id); !ok {
			t.Fatal("the slot did not restore")
		}
		restored.beginEventBatchLocked()
		if restored.resolving != nil {
			t.Error("beginEventBatchLocked left the restored slot set")
		}
	})
	// And the cleared state is what the next capture writes.
	raw, err := json.Marshal(restored.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(raw, &keys); err != nil {
		t.Fatal(err)
	}
	if _, present := keys["resolving"]; present {
		t.Error(`an empty slot was written as a "resolving" key`)
	}
}

// A file written before the field has no key, and restores with the
// slot empty: additive within v9.
func TestAFileWithoutAResolvingSlotRestoresEmpty(t *testing.T) {
	g := newRestorableGame(t)
	slot, _ := resolvingSpellSlot(g)
	g.WithWriteLock(func() { g.resolving = slot })

	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(raw, &keys); err != nil {
		t.Fatal(err)
	}
	if _, present := keys["resolving"]; !present {
		t.Fatal(`a set slot was not written as a "resolving" key`)
	}
	delete(keys, "resolving")
	old, err := json.Marshal(keys)
	if err != nil {
		t.Fatal(err)
	}
	var decoded GameSnapshot
	if err := json.Unmarshal(old, &decoded); err != nil {
		t.Fatal(err)
	}
	restored, err := decoded.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict of a file with no resolving key: %v", err)
	}
	if restored.resolving != nil {
		t.Errorf("a file with no resolving key restored a slot: %+v", restored.resolving)
	}
}
