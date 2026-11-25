package game

import (
	"testing"

	"github.com/google/uuid"
)

// retarget_change_all_test.go — RetargetChangeAll, "change the target or
// targets" (CR 115.7a, #6228): every target to another legal target, or
// none of them.

func castPairForChangeAll(t *testing.T, creatures int) (*Game, *Player, *StackItem, []TargetRefKindID) {
	t.Helper()
	g := newActiveGame(t)
	retargetSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	ids := make([]TargetRefKindID, creatures)
	for i := range ids {
		ids[i] = TargetRefKindID{ID: pushRetargetCreature(g, opp, string(rune('A'+i)))}
	}
	item := castTargeted(t, g, me, retargetPairOracle, []TargetRef{
		{Kind: TargetCard, ID: ids[0].ID, Slot: 0},
		{Kind: TargetCard, ID: ids[1].ID, Slot: 1},
	})
	return g, opp, item, ids
}

// TargetRefKindID is a test-local name for a seeded creature's id.
type TargetRefKindID struct{ ID uuid.UUID }

func offerChangeAll(t *testing.T, g *Game, item *StackItem, chooser *Player) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.OfferRetargetForEffect(RetargetOffer{
			ItemID: item.ID, Chooser: chooser.ID, Policy: RetargetChangeAll, Optional: true,
		}); err != nil {
			t.Fatalf("OfferRetargetForEffect: %v", err)
		}
	})
}

// Both slots asked, the item untouched until the last answer, then both
// change together.
func TestChangeAllChangesEveryTargetTogether(t *testing.T) {
	g, opp, item, ids := castPairForChangeAll(t, 4)
	a, b, c, d := ids[0].ID, ids[1].ID, ids[2].ID, ids[3].ID
	offerChangeAll(t, g, item, opp)

	first := findRetargetPrompt(g, opp.ID)
	if first == nil || first.RetargetSlot != 0 || first.PickTargetMin != 0 {
		t.Fatalf("first prompt = %+v, want an optional slot 0", first)
	}
	if hasUUID(first.PickTargetCards, a) {
		t.Error("slot 0 offers the target already in it")
	}
	if err := g.ResolveRetarget(first.ID, opp.ID, []TargetRef{{Kind: TargetCard, ID: c}}); err != nil {
		t.Fatalf("slot 0: %v", err)
	}
	if got := g.StackMeta[item.ID].Targets; got[0].ID != a {
		t.Fatal("a target changed before every slot was answered")
	}
	second := findRetargetPrompt(g, opp.ID)
	if second == nil || second.RetargetSlot != 1 || second.PickTargetMin != 1 {
		t.Fatalf("second prompt = %+v, want a mandatory slot 1", second)
	}
	if hasUUID(second.PickTargetCards, b) || hasUUID(second.PickTargetCards, c) {
		t.Error("slot 1 offers its own target or the other slot's new pick")
	}
	if err := g.ResolveRetarget(second.ID, opp.ID, nil); err == nil {
		t.Fatal("slot 1 was declined after the change was begun")
	}
	if err := g.ResolveRetarget(second.ID, opp.ID, []TargetRef{{Kind: TargetCard, ID: d}}); err != nil {
		t.Fatalf("slot 1: %v", err)
	}
	got := g.StackMeta[item.ID].Targets
	if len(got) != 2 || got[0].ID != c || got[1].ID != d {
		t.Errorf("targets = %+v, want [C D]", got)
	}
}

// Declining at the first slot changes nothing.
func TestChangeAllMayBeDeclined(t *testing.T) {
	g, opp, item, ids := castPairForChangeAll(t, 4)
	offerChangeAll(t, g, item, opp)
	first := findRetargetPrompt(g, opp.ID)
	if first == nil {
		t.Fatal("no prompt")
	}
	if err := g.ResolveRetarget(first.ID, opp.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if findRetargetPrompt(g, opp.ID) != nil {
		t.Error("a declined change kept asking")
	}
	if got := g.StackMeta[item.ID].Targets; got[0].ID != ids[0].ID || got[1].ID != ids[1].ID {
		t.Errorf("targets changed: %+v", got)
	}
}

// CR 115.7e judges only the final set: two creatures may swap.
func TestChangeAllMaySwapTwoTargets(t *testing.T) {
	g, opp, item, ids := castPairForChangeAll(t, 2)
	a, b := ids[0].ID, ids[1].ID
	offerChangeAll(t, g, item, opp)
	first := findRetargetPrompt(g, opp.ID)
	if first == nil || !hasUUID(first.PickTargetCards, b) {
		t.Fatalf("slot 0 = %+v, want B offered", first)
	}
	if err := g.ResolveRetarget(first.ID, opp.ID, []TargetRef{{Kind: TargetCard, ID: b}}); err != nil {
		t.Fatalf("slot 0: %v", err)
	}
	second := findRetargetPrompt(g, opp.ID)
	if second == nil || !hasUUID(second.PickTargetCards, a) {
		t.Fatalf("slot 1 = %+v, want A offered", second)
	}
	if err := g.ResolveRetarget(second.ID, opp.ID, []TargetRef{{Kind: TargetCard, ID: a}}); err != nil {
		t.Fatalf("slot 1: %v", err)
	}
	if got := g.StackMeta[item.ID].Targets; got[0].ID != b || got[1].ID != a {
		t.Errorf("targets = %+v, want [B A]", got)
	}
}

// CR 115.7a: a target with no other legal target means none change.
// With B and C gone, A is the only creature left: slot 0 can go nowhere
// else, so nothing is asked and nothing changes.
func TestChangeAllWithATargetThatCannotMoveChangesNothing(t *testing.T) {
	g, opp, item, ids := castPairForChangeAll(t, 3)
	g.WithWriteLock(func() {
		g.Battlefield.Remove(ids[2].ID)
		g.Battlefield.Remove(ids[1].ID)
	})
	offerChangeAll(t, g, item, opp)
	if p := findRetargetPrompt(g, opp.ID); p != nil {
		t.Fatalf("asked with nowhere to go: slot %d", p.RetargetSlot)
	}
	if got := g.StackMeta[item.ID].Targets; got[0].ID != ids[0].ID || got[1].ID != ids[1].ID {
		t.Errorf("a change was made: %+v", got)
	}
}

// One target: the walk is "change the target" (CR 115.7b's one slot).
func TestChangeAllOverOneTargetIsChangeOne(t *testing.T) {
	g := newActiveGame(t)
	retargetSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushRetargetCreature(g, opp, "A")
	b := pushRetargetCreature(g, opp, "B")
	item := castTargeted(t, g, me, retargetOneOracle, []TargetRef{{Kind: TargetCard, ID: a}})
	offerChangeAll(t, g, item, opp)
	p := findRetargetPrompt(g, opp.ID)
	if p == nil || p.RetargetPolicy != RetargetChangeOne {
		t.Fatalf("prompt = %+v, want a change-one walk", p)
	}
	if err := g.ResolveRetarget(p.ID, opp.ID, []TargetRef{{Kind: TargetCard, ID: b}}); err != nil {
		t.Fatalf("answer: %v", err)
	}
	if got := g.StackMeta[item.ID].Targets; got[0].ID != b {
		t.Errorf("targets = %+v, want [B]", got)
	}
}
