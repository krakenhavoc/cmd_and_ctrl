package game

import (
	"slices"
	"testing"
)

// two_kickers_test.go — #2153, CR 702.33b: "Kicker [A] and/or [B]"
// means "Kicker [A]" and "Kicker [B]" — two optional costs, both keyed
// KickerKey, each paid at most once. The announcement names indices,
// so the wire never needed to change; what this pins is that the
// validator, the price, the record, the readers and a copy all treat
// the two as separate payments of the same keyword.

const testTwoKickerOracle = "test-two-kickers"

func twoKickers() []AdditionalCost {
	return []AdditionalCost{
		{Optional: true, Key: KickerKey, ManaCost: "{R}", Label: "Kicker {R}"},
		{Optional: true, Key: KickerKey, ManaCost: "{2}{W}", Label: "Kicker {2}{W}"},
	}
}

// TestTwoKickerAnnouncementIsEitherBothOrNeither is CR 601.2b over the
// pair: any subset, each kicker at most once.
func TestTwoKickerAnnouncementIsEitherBothOrNeither(t *testing.T) {
	for _, tc := range []struct {
		claim   []int
		wantErr bool
	}{
		{nil, false},
		{[]int{0}, false},
		{[]int{1}, false},
		{[]int{0, 1}, false},
		{[]int{1, 0}, false},
		{[]int{0, 0}, true},
		{[]int{1, 1}, true},
		{[]int{0, 1, 1}, true},
		{[]int{2}, true},
	} {
		err := validateOptionalCostChoice(twoKickers(), tc.claim)
		if (err != nil) != tc.wantErr {
			t.Errorf("validateOptionalCostChoice(%v) = %v, wantErr %v", tc.claim, err, tc.wantErr)
		}
	}
}

// TestTwoKickersEachAddTheirOwnMana is CR 601.2f: each kicker paid
// joins the total, and only the ones paid.
func TestTwoKickersEachAddTheirOwnMana(t *testing.T) {
	base, err := ParseCost("{2}{G}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	for _, tc := range []struct {
		claim             []int
		generic, coloured int
	}{
		{nil, 2, 1},
		{[]int{0}, 2, 2},
		{[]int{1}, 4, 2},
		{[]int{0, 1}, 4, 3},
	} {
		got, err := AdditionalCostMana(base, nil, twoKickers(), tc.claim)
		if err != nil {
			t.Fatalf("AdditionalCostMana(%v): %v", tc.claim, err)
		}
		if got.Generic != tc.generic || len(got.Required) != tc.coloured {
			t.Errorf("claim %v: generic %d coloured %d, want %d and %d",
				tc.claim, got.Generic, len(got.Required), tc.generic, tc.coloured)
		}
	}
}

// TestTwoKickerReadersTellThePaymentsApart is every reader at once:
// the record (normalised), "the number of times it was kicked" (CR
// 702.33d — both is twice), and "kicked with its [A] kicker" (CR
// 702.33f) for each half, for a cost the card does not print, and on
// the permanent the spell becomes.
func TestTwoKickerReadersTellThePaymentsApart(t *testing.T) {
	for _, tc := range []struct {
		name     string
		claim    []int
		record   []int
		times    int
		withR    bool
		withW    bool
		modeKick bool
	}{
		{"neither", nil, nil, 0, false, false, false},
		{"first", []int{0}, []int{0}, 1, true, false, true},
		{"second", []int{1}, []int{1}, 1, false, true, true},
		{"both, announced backwards", []int{1, 0}, []int{0, 1}, 2, true, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newRestorableGame(t)
			stubOptionalCosts(t, testTwoKickerOracle, twoKickers())

			id, err := castOptional(t, g, "Two-Kicker Thing", "Creature — Test", testTwoKickerOracle, tc.claim)
			if err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			item := g.StackMeta[id]
			if item == nil {
				t.Fatal("no stack item")
			}
			if !slices.Equal(item.Paid.OptionalCosts, tc.record) {
				t.Errorf("record = %v, want %v", item.Paid.OptionalCosts, tc.record)
			}
			card := Card{OracleID: testTwoKickerOracle}
			if got := KickedTimesPaid(card, item.Paid.OptionalCosts); got != tc.times {
				t.Errorf("KickedTimesPaid = %d, want %d", got, tc.times)
			}
			if got := KickedWithPaid(card, item.Paid.OptionalCosts, "{R}"); got != tc.withR {
				t.Errorf("kicked with {R} = %v, want %v", got, tc.withR)
			}
			if got := KickedWithPaid(card, item.Paid.OptionalCosts, "{2}{W}"); got != tc.withW {
				t.Errorf("kicked with {2}{W} = %v, want %v", got, tc.withW)
			}
			if KickedWithPaid(card, item.Paid.OptionalCosts, "{W}") || KickedWithPaid(card, item.Paid.OptionalCosts, "") {
				t.Error("a kicker the card does not print reads as paid")
			}
			q := ModeCountQuery{OracleID: testTwoKickerOracle, OptionalCosts: item.Paid.OptionalCosts}
			if q.Kicked() != tc.modeKick {
				t.Errorf("ModeCountQuery.Kicked = %v, want %v", q.Kicked(), tc.modeKick)
			}

			// CR 400.7d: the permanent the spell becomes answers the same.
			permanent := Card{OracleID: testTwoKickerOracle, Provenance: CastProvenance{OptionalCosts: item.Paid.OptionalCosts}}
			if CardKickedTimes(permanent) != tc.times || CardKickedWith(permanent, "{R}") != tc.withR ||
				CardKickedWith(permanent, "{2}{W}") != tc.withW {
				t.Errorf("the permanent reads times=%d R=%v W=%v, want %d %v %v",
					CardKickedTimes(permanent), CardKickedWith(permanent, "{R}"), CardKickedWith(permanent, "{2}{W}"),
					tc.times, tc.withR, tc.withW)
			}
		})
	}
}

// TestACopyOfASpellKeepsWhichKickerWasPaid is CR 707.10b over the
// pair: a copy of a spell kicked with its second kicker was kicked
// with that one, once, and not with the first.
func TestACopyOfASpellKeepsWhichKickerWasPaid(t *testing.T) {
	g := newRestorableGame(t)
	stubOptionalCosts(t, testTwoKickerOracle, twoKickers())
	me := g.Seats[g.Turn.ActiveSeat]

	id, err := castOptional(t, g, "Two-Kicker Thing", "Sorcery", testTwoKickerOracle, []int{1})
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	g.mu.Lock()
	if err := g.CopySpellForEffect(id, me.ID, false, nil); err != nil {
		g.mu.Unlock()
		t.Fatalf("CopySpellForEffect: %v", err)
	}
	var copyItem *StackItem
	for itemID, it := range g.StackMeta {
		if itemID != id {
			copyItem = it
		}
	}
	var copyCard Card
	var found bool
	if copyItem != nil {
		copyCard, found = g.LookupCardForEffect(copyItem.ID)
	}
	g.mu.Unlock()
	if copyItem == nil || !found {
		t.Fatal("no copy was created")
	}
	if CatalogKey(copyCard) != testTwoKickerOracle {
		t.Fatalf("the copy's catalog key = %q; the readers below would prove nothing", CatalogKey(copyCard))
	}
	paid := copyItem.Paid.OptionalCosts
	if KickedTimesPaid(copyCard, paid) != 1 || !KickedWithPaid(copyCard, paid, "{2}{W}") || KickedWithPaid(copyCard, paid, "{R}") {
		t.Errorf("the copy reads times=%d W=%v R=%v, want 1 true false (CR 707.10b)",
			KickedTimesPaid(copyCard, paid), KickedWithPaid(copyCard, paid, "{2}{W}"), KickedWithPaid(copyCard, paid, "{R}"))
	}
}
