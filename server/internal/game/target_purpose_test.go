package game

import (
	"reflect"
	"testing"
)

// target_purpose_test.go — ADR 0126's amendment of 2026-10-08: target
// entries keep Purpose comparable, and a fused spell numbers the right
// half's entries past the left half's clauses, as its statement does.

func TestTargetEntriesKeepPurposeComparable(t *testing.T) {
	if !(Purpose{}).IsZero() {
		t.Fatal("the zero Purpose is not zero")
	}
	p := Purpose{Targets: ForTargets(TargetPurpose{Damage: 3})}
	if p.IsZero() {
		t.Error("a purpose with a target entry reads as zero")
	}
	if ForTargets() != nil {
		t.Error("ForTargets() with no entries is not nil")
	}
	var nilList *TargetPurposes
	if nilList.List() != nil {
		t.Error("a nil list has entries")
	}
}

func TestFusedPurposeRenumbersTheRightHalfsTargets(t *testing.T) {
	left := Purpose{Draws: 1, Targets: ForTargets(TargetPurpose{Slot: 0, Damage: 2})}
	right := Purpose{Targets: ForTargets(TargetPurpose{Slot: 0, Draws: 2}, TargetPurpose{Slot: 1, LifeLoss: 1})}
	got := left.plus(right, 1)
	want := []TargetPurpose{{Slot: 0, Damage: 2}, {Slot: 1, Draws: 2}, {Slot: 2, LifeLoss: 1}}
	if !reflect.DeepEqual(got.Targets.List(), want) {
		t.Errorf("fused targets = %+v, want %+v", got.Targets.List(), want)
	}
	if got.Draws != 1 {
		t.Errorf("fused draws = %d, want the left half's 1", got.Draws)
	}
	// Neither half's list is touched.
	if right.Targets.List()[0].Slot != 0 {
		t.Error("plus renumbered the right half's own list")
	}
	if none := (Purpose{Draws: 1}).plus(Purpose{Draws: 1}, 0); none.Targets != nil {
		t.Errorf("two halves with no entries fuse to %+v", none.Targets)
	}
}
