package game

import (
	"testing"

	"github.com/google/uuid"
)

// spell_entry_counters_mark_test.go — #4686: the effect writer of
// StackItem.PromisedCounters marks only the spell object it was told
// about, and only while it is on the stack.

func TestMarkSpellEntersWithCountersMarksOnlyThatObjectOnTheStack(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	floatMana(me, "GG")
	item := castForTest(t, g, me, "{G}", CastSpellParams{Strict: true})

	g.mu.Lock()
	defer g.mu.Unlock()
	epoch := g.cardObjectEpochLocked(item.ID)
	mark := PromisedCounter{Kind: CounterPlusOne, Count: 2, SourceName: "Test"}

	if g.MarkSpellEntersWithCountersForEffect(ObjectRef{ID: item.ID, Epoch: epoch + 1}, mark) {
		t.Error("a ref to a different object (another epoch) marked the spell")
	}
	if g.MarkSpellEntersWithCountersForEffect(ObjectRef{ID: uuid.New(), Epoch: 0}, mark) {
		t.Error("an ID on no stack was marked")
	}
	if g.MarkSpellEntersWithCountersForEffect(ObjectRef{ID: item.ID, Epoch: epoch}, PromisedCounter{Kind: CounterPlusOne}) {
		t.Error("a mark of zero counters was written")
	}
	if g.MarkSpellEntersWithCountersForEffect(ObjectRef{ID: item.ID, Epoch: epoch}, PromisedCounter{Count: 1}) {
		t.Error("a mark with no counter kind was written")
	}
	if len(item.PromisedCounters) != 0 {
		t.Fatalf("refused marks left %+v", item.PromisedCounters)
	}

	before := item.PromisedCounters
	if !g.MarkSpellEntersWithCountersForEffect(ObjectRef{ID: item.ID, Epoch: epoch}, mark) {
		t.Fatal("the spell on the stack was not marked")
	}
	if !g.MarkSpellEntersWithCountersForEffect(ObjectRef{ID: item.ID, Epoch: epoch}, mark) {
		t.Fatal("a second mark was refused")
	}
	if len(item.PromisedCounters) != 2 || item.PromisedCounters[0] != mark {
		t.Errorf("marks = %+v, want two of %+v", item.PromisedCounters, mark)
	}
	if len(before) != 0 {
		t.Error("the mark wrote into a shared backing array")
	}

	if err := g.CopySpellForEffect(item.ID, me.ID, false, nil); err != nil {
		t.Fatalf("copy: %v", err)
	}
	for id, it := range g.StackMeta {
		if !it.IsCopy {
			continue
		}
		if len(it.PromisedCounters) != 0 {
			t.Errorf("the copy inherited the original's marks: %+v", it.PromisedCounters)
		}
		if g.MarkSpellEntersWithCountersForEffect(ObjectRef{ID: id, Epoch: g.cardObjectEpochLocked(id)}, mark) {
			t.Error("a copy of a spell was marked")
		}
	}
}
