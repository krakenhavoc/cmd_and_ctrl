package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// gift_test.go — ADR 0089 (#1267), the engine half of gift (CR
// 702.174): the announcement's opponent, the record on the stack item
// and the permanent, and that both survive an undo and a snapshot. The
// card files in effects/gift_test.go prove the keyword end to end.

const testGiftOracle = "test-gift-spell"

func giftCostForTest(targets *TargetSpec) AdditionalCost {
	return AdditionalCost{Optional: true, Key: GiftKey, ChoosesOpponent: true, Label: "Gift a card", Targets: targets}
}

// castGiftForTest seeds a sorcery into the active seat's hand and casts
// it promising the gift to `to`.
func castGiftForTest(t *testing.T, g *Game, to uuid.UUID) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(Card{
		InstanceID: id, Name: "Gift Thing", TypeLine: "Sorcery", OracleID: testGiftOracle,
		ManaCost: "{1}", Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != StepPrecombatMain && g.Turn.Step != StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, CastSpellParams{OptionalCosts: []int{0}, GiftOpponent: to}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	return id
}

func TestGiftOpponentsAreTheOtherLivingSeats(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	g.Seats[2].Eliminated = true
	var got []uuid.UUID
	g.WithWriteLock(func() { got = g.GiftOpponentsLocked(me.ID) })
	want := []uuid.UUID{g.Seats[1].ID, g.Seats[3].ID}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("GiftOpponentsLocked = %v, want %v (not the caster, not a player who left)", got, want)
	}
}

// TestGiftPromiseSurvivesUndoAndSnapshot pins the record: the opponent
// the gift was promised to rides the stack item, a clone does not
// alias it, and a snapshot round trip brings it back.
func TestGiftPromiseSurvivesUndoAndSnapshot(t *testing.T) {
	g := newRestorableGame(t)
	stubOptionalCosts(t, testGiftOracle, []AdditionalCost{giftCostForTest(nil)})
	opp := g.Seats[1]
	id := castGiftForTest(t, g, opp.ID)
	if item := g.StackMeta[id]; item == nil || item.Paid.GiftOpponent != opp.ID {
		t.Fatalf("the stack item must record the promise to %v", opp.ID)
	}

	mid := g.Clone()
	g.WithWriteLock(func() { g.StackMeta[id].Paid.GiftOpponent = uuid.Nil })
	if mid.StackMeta[id].Paid.GiftOpponent != opp.ID {
		t.Fatal("the clone aliases the live record")
	}
	g.RestoreFrom(mid)
	if !g.StackMeta[id].Paid.GiftPromised() {
		t.Fatal("undo lost the promise")
	}

	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded GameSnapshot
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	restored, err := decoded.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if item := restored.StackMeta[id]; item == nil || item.Paid.GiftOpponent != opp.ID {
		t.Fatalf("the promise did not survive the snapshot")
	}
}

// TestGiftProvenanceSurvivesSnapshotAndClearsOnTheWayOut is CR 400.7d
// and CR 400.7 for the permanent half: the record is carried, and a
// permanent that leaves and comes back was not cast with a promise.
func TestGiftProvenanceSurvivesSnapshotAndClearsOnTheWayOut(t *testing.T) {
	g := newRestorableGame(t)
	owner, opp := g.Seats[0], g.Seats[1]
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id, Name: "Promised Permanent", TypeLine: "Creature — Raccoon",
		Owner: owner.ID, Controller: owner.ID,
		Provenance: CastProvenance{OptionalCosts: []int{0}, GiftOpponent: opp.ID},
	})
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded GameSnapshot
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	restored, err := decoded.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	found := false
	for _, c := range restored.Battlefield.Cards {
		if c.InstanceID == id {
			found = c.GiftPromised() && c.Provenance.GiftOpponent == opp.ID
		}
	}
	if !found {
		t.Fatal("Provenance.GiftOpponent did not survive the snapshot")
	}

	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(id); err != nil {
			t.Fatalf("bounce: %v", err)
		}
	})
	for _, c := range owner.Hand.Cards {
		if c.InstanceID == id && c.GiftPromised() {
			t.Error("the promise survived the permanent leaving the battlefield")
		}
	}
}

// TestGiftResolvedPermanentCarriesThePromise is the stamp: a resolving
// permanent spell hands its promise to the permanent it becomes.
func TestGiftResolvedPermanentCarriesThePromise(t *testing.T) {
	g := newRestorableGame(t)
	const oracle = "test-gift-creature"
	stubOptionalCosts(t, oracle, []AdditionalCost{giftCostForTest(nil)})
	me, opp := g.Seats[0], g.Seats[1]
	id := uuid.New()
	me.Hand.PushTop(Card{
		InstanceID: id, Name: "Gift Bear", TypeLine: "Creature — Bear", OracleID: oracle,
		ManaCost: "{1}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	for g.Turn.Step != StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{OptionalCosts: []int{0}, GiftOpponent: opp.ID}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	for i := 0; i < len(g.Seats) && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok || !g.Battlefield.Contains(id) {
		t.Fatal("the creature should have resolved onto the battlefield")
	}
	if c.Provenance.GiftOpponent != opp.ID {
		t.Errorf("Provenance.GiftOpponent = %v, want %v", c.Provenance.GiftOpponent, opp.ID)
	}
}

// TestTargetSpecUnderOptionalCostsSwapsOnlyWhenPaid is the clause
// rewrite on its own (CR 702.174m): the paid cost's clause, the
// printed one otherwise.
func TestTargetSpecUnderOptionalCostsSwapsOnlyWhenPaid(t *testing.T) {
	base := &TargetSpec{Mode: "creature", Label: "target creature spell"}
	wide := &TargetSpec{Mode: "spell", Label: "target spell"}
	costs := []AdditionalCost{
		{Optional: true, Key: KickerKey, ManaCost: "{1}"},
		giftCostForTest(wide),
	}
	if got := TargetSpecUnderOptionalCosts(base, costs, nil); got != base {
		t.Error("unpaid: the printed clause")
	}
	if got := TargetSpecUnderOptionalCosts(base, costs, []int{0}); got != base {
		t.Error("a cost with no rewrite leaves the clause alone")
	}
	if got := TargetSpecUnderOptionalCosts(base, costs, []int{1}); got != wide {
		t.Error("the promised gift's clause replaces the printed one")
	}
	if got := TargetSpecUnderOptionalCosts(nil, costs, []int{1}); got != wide {
		t.Error("a clause only the promise has (CR 702.174m)")
	}
}
