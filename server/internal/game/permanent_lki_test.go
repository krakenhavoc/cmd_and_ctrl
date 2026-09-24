package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// permanent_lki_test.go covers #1379: resolution-time last-known
// information for a permanent that has left the battlefield
// (CR 608.2h). The card-level proof is in the effects package (Cream of
// the Crop, Tribute to the World Tree, Warstorm Surge, Murderous Redcap,
// Claustrophobia).

// refOf names the permanent object `id` is right now.
func refOf(t *testing.T, g *Game, id uuid.UUID) ObjectRef {
	t.Helper()
	var ref ObjectRef
	var ok bool
	g.WithWriteLock(func() { ref, ok = g.PermanentRefForEffect(id) })
	if !ok {
		t.Fatalf("PermanentRefForEffect(%v): no object", id)
	}
	return ref
}

// readPermanent is PermanentForEffect under the write lock.
func readPermanent(g *Game, ref ObjectRef) (PermanentInfo, bool) {
	var info PermanentInfo
	var ok bool
	g.WithWriteLock(func() { info, ok = g.PermanentForEffect(ref) })
	return info, ok
}

// The headline: live while it is there — so a counter that arrived
// while the ability waited counts — and as it last existed once it has
// gone, counters included. A resolution a priority round after the
// removal is exactly what the CR 603.10 store could not answer: the
// harvest has long since deleted its entry.
func TestPermanentForEffectReadsLiveThenLastKnown(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushBear(g, me.ID)
	ref := refOf(t, g, bear)

	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(bear, CounterPlusOne, 2); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})
	live, ok := readPermanent(g, ref)
	if !ok || live.Left || live.Power != 4 || live.Toughness != 4 {
		t.Fatalf("live read = %+v, %v; want a 4/4 still on the battlefield", live, ok)
	}

	destroy(t, g, bear)
	if _, still := g.lastKnownBattlefield[bear]; still {
		t.Fatal("setup: the CR 603.10 entry should be gone after the harvest")
	}
	gone, ok := readPermanent(g, ref)
	if !ok || !gone.Left {
		t.Fatalf("after the destroy = %+v, %v; want last-known information", gone, ok)
	}
	if gone.Power != 4 || gone.Toughness != 4 || gone.Counters[CounterPlusOne] != 2 {
		t.Errorf("last-known = %d/%d with %v; want 4/4 with two +1/+1 counters", gone.Power, gone.Toughness, gone.Counters)
	}
	if gone.Controller != me.ID || !typeListHas(gone.Characteristic.Types, "creature") {
		t.Errorf("last-known controller %v types %v; want %v and a creature", gone.Controller, gone.Characteristic.Types, me.ID)
	}
}

// CR 400.7: the same card back on the battlefield is a new object. A ref
// taken for the first object must read the first object's record, never
// the impostor's live values — and a ref taken for the new one reads it
// live.
func TestPermanentForEffectNeverAnswersWithTheNewObject(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushBear(g, me.ID)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, CounterPlusOne, 1) })
	first := refOf(t, g, bear)

	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(bear); err != nil {
			t.Fatalf("BounceToHandForEffect: %v", err)
		}
	})
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneHand, Owner: me.ID}, ZoneRef{Kind: ZoneBattlefield}, bear); err != nil {
		t.Fatalf("MoveCardByID back: %v", err)
	}
	second := refOf(t, g, bear)
	if second == first {
		t.Fatal("setup: the returned card should be a new object")
	}

	old, ok := readPermanent(g, first)
	if !ok || !old.Left || old.Power != 3 {
		t.Errorf("the first object = %+v, %v; want its last-known 3 power", old, ok)
	}
	now, ok := readPermanent(g, second)
	if !ok || now.Left || now.Power != 2 {
		t.Errorf("the new object = %+v, %v; want a live, counterless 2 power", now, ok)
	}
}

// One card, two departures in a turn, two objects: each ref reads its
// own record, and PermanentRefForEffect on a card that is not on the
// battlefield names the LAST object it was.
func TestPermanentLKIKeepsEveryDepartedObject(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushBear(g, me.ID)
	first := refOf(t, g, bear)
	hand := ZoneRef{Kind: ZoneHand, Owner: me.ID}
	bf := ZoneRef{Kind: ZoneBattlefield}

	if err := g.MoveCardByID(bf, hand, bear); err != nil {
		t.Fatalf("first exit: %v", err)
	}
	if err := g.MoveCardByID(hand, bf, bear); err != nil {
		t.Fatalf("return: %v", err)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, CounterPlusOne, 5) })
	second := refOf(t, g, bear)
	if err := g.MoveCardByID(bf, hand, bear); err != nil {
		t.Fatalf("second exit: %v", err)
	}

	if a, ok := readPermanent(g, first); !ok || a.Power != 2 {
		t.Errorf("first object = %+v, %v; want its own 2 power", a, ok)
	}
	if b, ok := readPermanent(g, second); !ok || b.Power != 7 {
		t.Errorf("second object = %+v, %v; want its own 7 power", b, ok)
	}
	if last := refOf(t, g, bear); last != second {
		t.Errorf("the card in hand names %+v, want the last object it was, %+v", last, second)
	}
}

// An Aura's "enchanted creature" after the Aura has gone reads what it
// was attached to.
func TestPermanentLKIRemembersTheAttachment(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	host := pushBear(g, me.ID)
	aura := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: aura, Name: "Aura", TypeLine: "Enchantment — Aura",
		Owner: me.ID, Controller: me.ID,
		AttachedTo: TargetRef{Kind: TargetCard, ID: host},
	})
	ref := refOf(t, g, aura)
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(aura); err != nil {
			t.Fatalf("ExileCardForEffect: %v", err)
		}
	})
	got, ok := readPermanent(g, ref)
	if !ok || got.AttachedTo.Kind != TargetCard || got.AttachedTo.ID != host {
		t.Errorf("the departed Aura = %+v, %v; want it attached to %v", got, ok, host)
	}
}

// The trigger context names the object the event was about: for an exit
// that is the object that LEFT, not the card's new graveyard epoch — so
// an LTB trigger's resolution reads the right record.
func TestObjectSnapshotNamesTheDepartedObject(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushBear(g, me.ID)
	ref := refOf(t, g, bear)
	g.WithWriteLock(func() {
		g.battlefieldExitLocked(bear)
		if _, err := MoveCard(g.Battlefield, me.Graveyard, bear); err != nil {
			t.Fatalf("MoveCard: %v", err)
		}
		snap := g.objectSnapshotLocked(bear)
		if snap.Ref() != ref {
			t.Errorf("snapshot names %+v, want the departed object %+v", snap.Ref(), ref)
		}
		if info, ok := g.PermanentForEffect(snap.Ref()); !ok || !info.Left {
			t.Errorf("the snapshot's ref reads %+v, %v; want the record", info, ok)
		}
	})
}

// Undo rewinds the record with the permanent; the persisted snapshot
// carries it, so a restore mid-stack still has the answer.
func TestPermanentLKIRidesCloneAndSnapshot(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushBear(g, me.ID)
	ref := refOf(t, g, bear)
	before := g.Clone()

	destroy(t, g, bear)
	after := g.Clone()

	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	restored, err := snap.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if info, ok := readPermanent(restored, ref); !ok || !info.Left || info.Power != 2 {
		t.Errorf("the restored game reads %+v, %v; want the record", info, ok)
	}

	g.WithWriteLock(func() { g.RestoreFrom(before) })
	if info, ok := readPermanent(g, ref); !ok || info.Left {
		t.Errorf("undone to before the destroy: %+v, %v; want the live bear", info, ok)
	}
	g.WithWriteLock(func() { g.RestoreFrom(after) })
	if info, ok := readPermanent(g, ref); !ok || !info.Left {
		t.Errorf("undone to after the destroy: %+v, %v; want the record", info, ok)
	}
	// A second departure on the restored game must not write through
	// to the undo point it came from.
	other := pushBear(g, me.ID)
	destroy(t, g, other)
	after.mu.RLock()
	_, leaked := after.lastKnownPermanents[other]
	after.mu.RUnlock()
	if leaked {
		t.Error("the restored game shares its record with the undo point")
	}
}

// Nothing can name a permanent from an earlier turn: whatever referred
// to it was on the stack, and the stack is empty when a turn ends.
func TestPermanentLKIIsClearedAtTheTurnBoundary(t *testing.T) {
	g := newActiveGame(t)
	bear := pushBear(g, g.Seats[0].ID)
	ref := refOf(t, g, bear)
	destroy(t, g, bear)
	g.WithWriteLock(func() {
		g.onTurnBeganLocked()
		if len(g.lastKnownPermanents) != 0 {
			t.Errorf("%d records outlived their turn", len(g.lastKnownPermanents))
		}
	})
	if _, ok := readPermanent(g, ref); ok {
		t.Error("a record from the previous turn was read")
	}
}

// CR 800.4a: an object that has left the game has nothing left to know.
func TestPermanentLKIIsForgottenWhenItsOwnerLeaves(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	them := g.Seats[1]
	bear := pushBear(g, them.ID)
	ref := refOf(t, g, bear)
	destroy(t, g, bear)
	if err := g.Concede(them.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if _, ok := readPermanent(g, ref); ok {
		t.Error("the departed player's creature still has a record")
	}
}
