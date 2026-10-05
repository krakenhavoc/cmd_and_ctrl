package game

import (
	"slices"
	"testing"
)

// mana_no_generic_test.go — #2170: "this mana can't be spent to pay
// generic mana costs" (Jegantha, the Wellspring), ManaRestrictNoGeneric.
// The card half is cards/effects/jegantha_the_wellspring_test.go; the
// enumerator's and the view's agreement with the payment are in
// internal/legal and internal/protocol.

var noGen = []string{ManaRestrictNoGeneric}

// wubrgNoGeneric is what one Jegantha tap adds.
func wubrgNoGeneric() ManaPool {
	var p ManaPool
	for _, c := range []string{"W", "U", "B", "R", "G"} {
		p = append(p, ManaToken{Color: c, Restrictions: noGen})
	}
	return p
}

var castCtx = ManaSpendContext{Purpose: SpendPurposeCast}

func TestNoGenericManaPaysColoredSymbolsButNotGeneric(t *testing.T) {
	pool := wubrgNoGeneric()
	for _, tc := range []struct {
		cost string
		x    int
		want bool
	}{
		{"{W}{U}{B}{R}{G}", 0, true},
		{"{W}{U}", 0, true},
		{"{4}", 0, false},
		{"{1}", 0, false},
		{"{3}{W}", 0, false},
		{"{X}{R}", 0, true}, // X = 0
		{"{X}{R}", 1, false},
		{"{W/U}{B/R}", 0, true},          // hybrid
		{"{B/P}{R/P}", 0, true},          // Phyrexian
		{"{2/W}{2/U}", 0, true},          // two-mana hybrid, coloured half
		{"{W}{U}{B}{R}{G}{G}", 0, false}, // a sixth coloured symbol has no mana
	} {
		if got := pool.CanPayFor(spendOnlyCost(t, tc.cost), tc.x, castCtx); got != tc.want {
			t.Errorf("WUBRG (no generic) paying %s with X=%d: %v, want %v", tc.cost, tc.x, got, tc.want)
		}
	}
}

// Mixed with other mana: the ordinary mana pays the generic part and the
// restricted mana pays the coloured part, and the solver burns the
// restricted mana first so the ordinary mana is the one left over.
func TestNoGenericManaMixedWithOrdinaryMana(t *testing.T) {
	pool := append(wubrgNoGeneric(), ManaToken{Color: "R"}, ManaToken{Color: "C"})
	cost := spendOnlyCost(t, "{2}{W}{U}")
	if !pool.CanPayFor(cost, 0, castCtx) {
		t.Fatal("two ordinary mana and W U off Jegantha did not pay {2}{W}{U}")
	}
	spent, ok := pool.SpendManaFor(cost, 0, castCtx)
	if !ok {
		t.Fatal("SpendManaFor refused a payable cost")
	}
	if len(spent) != 4 {
		t.Fatalf("spent %d tokens, want 4", len(spent))
	}
	var plain, restricted int
	for _, tok := range spent {
		if tok.noGeneric() {
			restricted++
		} else {
			plain++
		}
	}
	if plain != 2 || restricted != 2 {
		t.Errorf("spent %d ordinary and %d restricted, want 2 and 2", plain, restricted)
	}
	// One ordinary mana short of the generic: the three leftover
	// restricted mana do not make up the difference.
	short := append(wubrgNoGeneric(), ManaToken{Color: "R"})
	if short.CanPayFor(cost, 0, castCtx) {
		t.Error("{2}{W}{U} paid with one ordinary mana and three spare restricted mana")
	}
}

func TestNoGenericManaAndXSpells(t *testing.T) {
	cost := spendOnlyCost(t, "{X}{R}{R}")
	pool := ManaPool{
		{Color: "R", Restrictions: noGen}, {Color: "R", Restrictions: noGen},
		{Color: "G", Restrictions: noGen}, {Color: "G", Restrictions: noGen},
	}
	if !pool.CanPayFor(cost, 0, castCtx) {
		t.Error("{R}{R} (X=0) refused")
	}
	if pool.CanPayFor(cost, 2, castCtx) {
		t.Error("X=2 paid with restricted green mana")
	}
	pool = append(pool, ManaToken{Color: "G"}, ManaToken{Color: "G"})
	if !pool.CanPayFor(cost, 2, castCtx) {
		t.Error("X=2 refused with two ordinary mana")
	}
	if pool.CanPayFor(cost, 3, castCtx) {
		t.Error("X=3 paid with two ordinary mana")
	}
}

// Commander tax and every other cost increase are generic mana added to
// the cost (CR 903.8), which is exactly the part this mana can't pay.
func TestNoGenericManaCannotPayTax(t *testing.T) {
	cost := spendOnlyCost(t, "{W}{U}{B}{R}{G}")
	pool := wubrgNoGeneric()
	if !pool.CanPayFor(cost, 0, castCtx) {
		t.Fatal("untaxed five-colour commander refused")
	}
	taxed := cost
	taxed.Generic += 2
	if pool.CanPayFor(taxed, 0, castCtx) {
		t.Error("a commander taxed by {2} was paid with restricted mana")
	}
	pool = append(pool, ManaToken{Color: "C"}, ManaToken{Color: "C"})
	if !pool.CanPayFor(taxed, 0, castCtx) {
		t.Error("the taxed commander was refused two ordinary mana for the tax")
	}
}

func TestNoGenericManaStillObeysItsOtherTags(t *testing.T) {
	pool := ManaPool{{Color: "W", Restrictions: []string{ManaRestrictActivate, ManaRestrictNoGeneric}}}
	w := spendOnlyCost(t, "{W}")
	if pool.CanPayFor(w, 0, castCtx) {
		t.Error("an activation-only token paid for a cast")
	}
	if !pool.CanPayFor(w, 0, ManaSpendContext{Purpose: SpendPurposeActivate}) {
		t.Error("an activation-only token refused an activation")
	}
	if pool.CanPayFor(w, 0, ManaSpendContext{}) {
		t.Error("an unknown purpose paid with a restricted token")
	}
}

// A distinct-colours spend (converge) must not reach for restricted mana
// to pay its generic half either.
func TestNoGenericManaIsSkippedByTheDistinctColorsStrategy(t *testing.T) {
	pool := append(wubrgNoGeneric(), ManaToken{Color: "G"})
	cost := spendOnlyCost(t, "{1}")
	spent, ok := pool.SpendManaForWith(cost, 0, castCtx, SpendDistinctColors)
	if !ok || len(spent) != 1 || spent[0].noGeneric() {
		t.Fatalf("spent %+v ok=%v, want the one ordinary G", spent, ok)
	}
	only := wubrgNoGeneric()
	if _, ok := only.SpendManaForWith(cost, 0, castCtx, SpendDistinctColors); ok {
		t.Error("distinct-colours strategy paid {1} with restricted mana")
	}
}

func TestNoGenericManaMissingForNamesTheGenericShortfall(t *testing.T) {
	pool := wubrgNoGeneric()
	got := pool.MissingFor(spendOnlyCost(t, "{2}{W}"), 0, castCtx)
	if !slices.Equal(got, []string{"{1}", "{1}"}) {
		t.Errorf("missing %v, want the two generic symbols", got)
	}
	if got := pool.MissingFor(spendOnlyCost(t, "{W}{U}"), 0, castCtx); got != nil {
		t.Errorf("missing %v for a payable cost", got)
	}
}

// "Spend mana as though it were mana of any color" changes the colour,
// not the generic restriction (CR 609.4b): a widened coloured symbol is
// still a coloured symbol and takes the restricted mana, a generic one
// still does not.
func TestNoGenericManaUnderAnAnyColorGrant(t *testing.T) {
	g, me, _ := spendGrantTable(t)
	pushSpendGrant(g, me, anyColorYouOracle)
	ctx := ManaSpendContext{Purpose: SpendPurposeCast}
	asPaid := func(s string) ParsedCost {
		var out ParsedCost
		cost := spendOnlyCost(t, s)
		g.ReadSnapshot(func() { out = g.costAsPaidByLocked(me.ID, ctx, cost, 0) })
		return out
	}
	offColour := ManaPool{
		{Color: "R", Restrictions: noGen}, {Color: "R", Restrictions: noGen},
		{Color: "R", Restrictions: noGen},
	}
	if !offColour.CanPayFor(asPaid("{U}{U}{B}"), 0, ctx) {
		t.Error("restricted red mana refused coloured symbols under the grant")
	}
	if offColour.CanPayFor(asPaid("{U}{U}{2}"), 0, ctx) {
		t.Error("restricted red mana paid a generic {2} under the grant")
	}
	if offColour.CanPayFor(asPaid("{3}"), 0, ctx) {
		t.Error("restricted red mana paid {3} under the grant")
	}
}

// A coloured symbol a cast permission folded into the generic demand
// (FoldedColored) is still a coloured symbol; the rest of the generic
// demand is not.
func TestNoGenericManaPaysFoldedColoredSymbolsOnly(t *testing.T) {
	folded := ParsedCost{Generic: 3, FoldedColored: 2}
	pool := ManaPool{
		{Color: "W", Restrictions: noGen}, {Color: "U", Restrictions: noGen},
		{Color: "B", Restrictions: noGen},
	}
	if pool.CanPayFor(folded, 0, castCtx) {
		t.Error("three restricted mana paid a cost with one genuine generic")
	}
	pool = append(pool, ManaToken{Color: "R"})
	if !pool.CanPayFor(folded, 0, castCtx) {
		t.Error("two restricted and one ordinary mana refused {1} plus two folded symbols")
	}
}

// The auto-tapper's pool credit: restricted mana is credited to the
// coloured symbols and never to the generic, so the planner is asked
// for exactly the generic part.
func TestNoGenericManaPoolAndAutoTapAgree(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	cost := spendOnlyCost(t, "{2}{W}{U}")
	me.ManaPool = wubrgNoGeneric()

	if _, _, ok := g.autoTapTopUpLocked(me.ID, cost, 0, castCtx, nil, 0); ok {
		t.Fatal("the top-up planned a {2} generic off an empty battlefield and spare restricted mana")
	}
	pushBattlefieldForTest(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	if g.AutoTapTopUpForEffectExcluding(me.ID, cost, 0, castCtx, nil) {
		t.Error("one Mountain plus three spare restricted mana planned {2}{W}{U}")
	}
	pushBattlefieldForTest(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	if !g.AutoTapTopUpForEffectExcluding(me.ID, cost, 0, castCtx, nil) {
		t.Error("two Mountains and W U off the pool did not plan {2}{W}{U}")
	}
	shorts := poolShortfalls(me.ManaPool, cost, 0, castCtx)
	if len(shorts) != 1 || shorts[0].Generic != 2 || len(shorts[0].Required) != 0 {
		t.Errorf("shortfalls %+v, want one asking for {2} and no coloured symbol", shorts)
	}
	// And with no pool at all the payment and the plan agree it is the
	// pool that matters: the spell is payable by the engine's own solver.
	if me.ManaPool.CanPayFor(cost, 0, castCtx) {
		t.Error("the pool alone paid {2}{W}{U}")
	}
}

// ManaRestrictAnyOf cannot hold the symbol-level tag: the alternative
// would silently stop restricting.
func TestNoGenericTagCannotSitInsideAnyOf(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("ManaRestrictAnyOf accepted ManaRestrictNoGeneric")
		}
	}()
	ManaRestrictAnyOf([]string{ManaRestrictCast, ManaRestrictNoGeneric})
}

// The tag rides the token and the token rides the snapshot.
func TestNoGenericTokenSurvivesASnapshotRoundTrip(t *testing.T) {
	g := newRestorableGame(t)
	me := g.Seats[0]
	me.ManaPool = wubrgNoGeneric()
	_, restored := roundTrip(t, g)
	pool := restored.Seats[0].ManaPool
	if len(pool) != 5 {
		t.Fatalf("restored pool %v", pool)
	}
	if pool.CanPayFor(spendOnlyCost(t, "{1}"), 0, castCtx) {
		t.Error("the restored token paid generic mana")
	}
	if !pool.CanPayFor(spendOnlyCost(t, "{W}{U}{B}{R}{G}"), 0, castCtx) {
		t.Error("the restored token refused its coloured symbols")
	}
}
