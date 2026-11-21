package game

import (
	"testing"

	"github.com/google/uuid"
)

// s82_gap_b156_test.go — GAP batch 156 (#5948): the asking item of a keyed
// prompt carries the split of a divided clause, so Kuldotha Flamefiend's "if
// you do" can deal the damage as it was announced.

func TestAskedItemCarriesTheDividedSplit(t *testing.T) {
	a, b := uuid.New(), uuid.New()
	item := &StackItem{ID: uuid.New(), Distribution: map[uuid.UUID]int{a: 1, b: 3}}
	asked := NewAskedItem(item)
	if asked.Distribution[a] != 1 || asked.Distribution[b] != 3 {
		t.Fatalf("asked item split = %v, want 1 and 3", asked.Distribution)
	}
	// A copy, not an alias: editing the item or the rebuilt item leaves the prompt's split alone.
	item.Distribution[a] = 9
	rebuilt := asked.StackItem()
	rebuilt.Distribution[b] = 9
	if asked.Distribution[a] != 1 || asked.Distribution[b] != 3 {
		t.Errorf("the asked item's split changed through a copy: %v", asked.Distribution)
	}
	if copyAskedItem(asked).Distribution[b] != 3 {
		t.Error("copyAskedItem dropped the split")
	}
	if NewAskedItem(&StackItem{ID: uuid.New()}).Distribution != nil {
		t.Error("an undivided item carries a split")
	}
}
