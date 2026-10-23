package game

import (
	"slices"
	"testing"
)

// spend_only_distinct_test.go — #2575, ADR 0040's amendment of
// 2026-10-22: "no more than one mana of each color may be spent this
// way" (Emblazoned Golem). The cards that prove it are in
// cards/effects/mana_caps_and_alt_sources_test.go.

func distinctXCost(t *testing.T) ParsedCost {
	t.Helper()
	cost, err := ParseCost("{X}{2}")
	if err != nil {
		t.Fatal(err)
	}
	cost.SpendOnly = &ManaSpendOnly{Colors: []string{"W", "U", "B", "R", "G"}, XOnly: true, Distinct: true}
	return cost
}

// The fold writes one single-colour requirement per restricted mana,
// from the clause's colours in order, and an unpayable one past five.
func TestDistinctFoldIsOneRequirementPerColour(t *testing.T) {
	cost := distinctXCost(t)
	got := cost.withSpendColors([]string{"U", "R"}).foldSpendOnly(2)
	if got.XSlots != 0 || got.SpendOnly != nil || got.folded.n != 0 || got.Generic != 2 {
		t.Fatalf("folded = %+v, want {2} and two listed requirements", got)
	}
	var opts []string
	for _, r := range got.Required {
		if len(r.Options) != 1 {
			t.Fatalf("requirement %+v names %d colours, want one", r, len(r.Options))
		}
		opts = append(opts, r.Options[0])
	}
	if !slices.Equal(opts, []string{"U", "R"}) {
		t.Errorf("requirements = %v, want [U R]", opts)
	}
	six := cost.foldSpendOnly(6)
	if len(six.Required) != 6 || !unpayableByAnyMana(six.Required[5]) {
		t.Errorf("X = 6 folded to %+v, want five colours and one requirement no mana pays", six.Required)
	}
}

// The choices are every subset of the clause's colours of the size the
// X needs, in WUBRG order; none when there is nothing to choose.
func TestDistinctSpendChoices(t *testing.T) {
	cost := distinctXCost(t)
	if got := cost.distinctSpendChoices(2); len(got) != 10 || !slices.Equal(got[0], []string{"W", "U"}) || !slices.Equal(got[9], []string{"R", "G"}) {
		t.Errorf("X = 2 choices = %v, want the ten pairs from [W U] to [R G]", got)
	}
	if got := cost.distinctSpendChoices(4); len(got) != 5 {
		t.Errorf("X = 4 choices = %v, want five", got)
	}
	for _, x := range []int{0, 5, 6} {
		if got := cost.distinctSpendChoices(x); got != nil {
			t.Errorf("X = %d choices = %v, want none", x, got)
		}
	}
	// #2701: a reduction that reaches X leaves less of it to restrict,
	// so X = 6 with one of it reduced away is five mana, one of each.
	reduced := cost
	reduced.Generic, reduced.XReduced = 0, 1
	if got := reduced.distinctSpendChoices(6); got != nil {
		t.Errorf("X = 6 less one choices = %v, want none", got)
	}
	folded := reduced.foldSpendOnly(6).Required
	if len(folded) != 5 || slices.ContainsFunc(folded, unpayableByAnyMana) {
		t.Errorf("X = 6 less one folded to %+v, want five payable colours", folded)
	}
	plain := cost
	plain.SpendOnly = &ManaSpendOnly{Colors: []string{"W", "U"}, XOnly: true}
	if got := plain.distinctSpendChoices(1); got != nil {
		t.Errorf("a clause without the cap chose %v", got)
	}
}

// The payment picks the subset the pool can pay: {B}{G} floating pays
// X = 2 as black and green, which is the eighth pair in WUBRG order.
func TestDistinctColoursAsPaidPicksWhatThePoolPays(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	p.ManaPool = ManaPool{{Color: "B"}, {Color: "G"}, {Color: "C"}, {Color: "C"}}
	var paid ParsedCost
	g.ReadSnapshot(func() {
		paid = g.costAsPaidByLocked(p.ID, ManaSpendContext{Purpose: SpendPurposeCast}, distinctXCost(t), 2)
	})
	var opts []string
	for _, r := range paid.Required {
		opts = append(opts, r.Options...)
	}
	if !slices.Equal(opts, []string{"B", "G"}) {
		t.Errorf("chose %v, want [B G]", opts)
	}
	if !p.ManaPool.CanPayFor(paid, 2, ManaSpendContext{Purpose: SpendPurposeCast}) {
		t.Error("the chosen colours are not payable from the pool")
	}
	// Two black is not one of each: no subset is payable, and the
	// first is returned for the shortfall.
	p.ManaPool = ManaPool{{Color: "B"}, {Color: "B"}, {Color: "C"}, {Color: "C"}}
	g.ReadSnapshot(func() {
		paid = g.costAsPaidByLocked(p.ID, ManaSpendContext{Purpose: SpendPurposeCast}, distinctXCost(t), 2)
	})
	if p.ManaPool.CanPayFor(paid, 2, ManaSpendContext{Purpose: SpendPurposeCast}) {
		t.Error("{B}{B}{C}{C} paid a distinct-colours X = 2")
	}
}
