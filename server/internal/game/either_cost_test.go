package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// either_cost_test.go — ADR 0100 §2: the pure halves of an either/or
// additional cost — which cost an announcement pays, what it adds to
// the price, what life the plan pays — and the record's copy rules.

func eitherForTest() *AdditionalCost {
	return &AdditionalCost{
		Label: "Discard a card or pay {5}",
		Either: []AdditionalCost{
			{DiscardCards: 1, Key: "discard", Label: "Discard a card"},
			{ManaCost: "{5}", Key: "mana", Label: "Pay {5}"},
		},
	}
}

func branchIdx(i int) *int { return &i }

func TestChosenAdditionalCost(t *testing.T) {
	either := eitherForTest()
	plain := &AdditionalCost{DiscardCards: 1}
	for _, tc := range []struct {
		why     string
		cost    *AdditionalCost
		branch  *int
		wantKey string
		wantErr bool
	}{
		{"branch 0", either, branchIdx(0), "discard", false},
		{"branch 1", either, branchIdx(1), "mana", false},
		{"no branch on a branched cost", either, nil, "", true},
		{"out of range", either, branchIdx(2), "", true},
		{"negative", either, branchIdx(-1), "", true},
		{"a branch on a plain cost", plain, branchIdx(0), "", true},
		{"a branch on no cost", nil, branchIdx(0), "", true},
	} {
		got, err := ChosenAdditionalCost(tc.cost, tc.branch)
		if (err != nil) != tc.wantErr {
			t.Errorf("%s: err = %v", tc.why, err)
			continue
		}
		if err != nil {
			if !errors.Is(err, ErrCostBranch) || !errors.Is(err, ErrInvalidParam) {
				t.Errorf("%s: err = %v, want ErrCostBranch wrapping ErrInvalidParam", tc.why, err)
			}
			continue
		}
		if got.Key != tc.wantKey {
			t.Errorf("%s: key %q, want %q", tc.why, got.Key, tc.wantKey)
		}
	}
	// A plain cost with no branch is returned as it is.
	if got, err := ChosenAdditionalCost(plain, nil); err != nil || got != plain {
		t.Fatalf("plain cost: %v, %v", got, err)
	}
}

// The chosen branch's mana is summed with the optional costs' mana, as
// one plan (ADR 0100 §2 replaces AddOptionalCostMana).
func TestAdditionalCostManaSumsTheBranchAndTheKicker(t *testing.T) {
	base, _ := ParseCost("{R}")
	branch := &eitherForTest().Either[1]
	kicker := []AdditionalCost{{Optional: true, Key: KickerKey, ManaCost: "{2}{R}"}}
	got, err := AdditionalCostMana(base, branch, kicker, []int{0})
	if err != nil {
		t.Fatalf("AdditionalCostMana: %v", err)
	}
	if got.ManaValue() != 9 {
		t.Fatalf("total %s (%d), want mana value 9", got.String(), got.ManaValue())
	}
	// The discard branch adds nothing.
	got, _ = AdditionalCostMana(base, &eitherForTest().Either[0], nil, nil)
	if got.ManaValue() != 1 {
		t.Fatalf("discard branch total %s, want {R}", got.String())
	}
}

// planLife sums every entry's fixed life and the announced X of a
// "pay X life" entry, so the CR 119.4 check and the payer agree.
func TestPlanLife(t *testing.T) {
	plan := []costPayment{
		{cost: AdditionalCost{PayLife: 3}, index: -1},
		{cost: AdditionalCost{PayLifeX: true}, index: 0},
	}
	if got := planLife(plan, 4); got != 7 {
		t.Fatalf("planLife = %d, want 7", got)
	}
	if got := planLife(nil, 4); got != 0 {
		t.Fatalf("planLife(nil) = %d", got)
	}
}

// The record: CostBranch is the index plus one, Discarded the named
// cards; both deep-copy and both count against IsZero.
func TestPaidCostBranchAndDiscardRecord(t *testing.T) {
	card := uuid.New()
	paid := paidWithBranchAndDiscards(PaidCost{}, branchIdx(0), []uuid.UUID{card})
	if paid.CostBranch != 1 || len(paid.Discarded) != 1 || paid.Discarded[0] != card {
		t.Fatalf("paid = %+v", paid)
	}
	if paid.IsZero() {
		t.Fatal("a record with a branch and a discard is zero")
	}
	clone := clonePaidCost(paid)
	clone.Discarded[0] = uuid.New()
	if paid.Discarded[0] != card {
		t.Fatal("clonePaidCost aliased Discarded")
	}
	if (PaidCost{CostBranch: 2}).IsZero() {
		t.Fatal("a record with only a branch is zero")
	}
	if !paidWithBranchAndDiscards(PaidCost{}, nil, nil).IsZero() {
		t.Fatal("no branch and no discard should stay zero")
	}
}
