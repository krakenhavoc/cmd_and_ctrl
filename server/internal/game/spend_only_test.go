package game

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
)

// spend_only_test.go — #1600, ADR 0040's 2026-10-03 amendment: a cost
// that says which mana may pay it ("Spend only mana of the chosen color
// to activate this ability", "Spend only black mana on X"), and the
// "monocolored" / "multicolored" mana tags. The card-side half — Throne
// of Eldraine, Crypt Rats, Crimson Hellkite, Pillar of the Paruns,
// Obsidian Obelisk — is cards/effects/spend_only_cards_test.go, which
// also holds the view / enumerator / payment agreement tests.

func requirementOptions(c ParsedCost) [][]string {
	out := make([][]string, 0, len(c.Required))
	for _, r := range c.Required {
		out = append(out, r.Options)
	}
	return out
}

func spendOnlyCost(t *testing.T, s string) ParsedCost {
	t.Helper()
	c, err := ParseCost(s)
	if err != nil {
		t.Fatalf("ParseCost(%q): %v", s, err)
	}
	return c
}

// The whole-cost form: Throne's {3} is three white symbols, and the
// cost no longer owes any generic.
func TestSpendOnlyFoldsAWholeCostIntoItsColor(t *testing.T) {
	c := spendOnlyCost(t, "{3}")
	c.SpendOnly = &ManaSpendOnly{Colors: []string{"W"}}
	got := c.foldSpendOnly(0)
	if got.Generic != 0 || got.XSlots != 0 || got.SpendOnly != nil {
		t.Fatalf("folded cost kept generic %d / X %d / restriction %v", got.Generic, got.XSlots, got.SpendOnly)
	}
	if want := [][]string{{"W"}, {"W"}, {"W"}}; !reflect.DeepEqual(requirementOptions(got), want) {
		t.Errorf("folded requirements %v, want %v", requirementOptions(got), want)
	}
	if got.String() != "{W}{W}{W}" || c.String() != "{3}" {
		t.Errorf("rendered %q from %q — the fold must not touch the cost it was given", got.String(), c.String())
	}
}

// "On X" restricts XSlots*x and nothing else: {1}{X} with X=2 owes one
// generic and two black.
func TestSpendOnlyOnXFoldsOnlyTheX(t *testing.T) {
	c := spendOnlyCost(t, "{1}{X}")
	c.SpendOnly = &ManaSpendOnly{Colors: []string{"B"}, XOnly: true}
	got := c.foldSpendOnly(2)
	if got.Generic != 1 || got.XSlots != 0 {
		t.Fatalf("generic %d / X slots %d, want 1 / 0", got.Generic, got.XSlots)
	}
	if want := [][]string{{"B"}, {"B"}}; !reflect.DeepEqual(requirementOptions(got), want) {
		t.Errorf("requirements %v, want %v", requirementOptions(got), want)
	}
	// X = 0 restricts nothing and owes nothing for X.
	if zero := c.foldSpendOnly(0); zero.Generic != 1 || len(zero.Required) != 0 {
		t.Errorf("X=0 folded to %v", zero)
	}
}

// A whole-cost restriction narrows a coloured symbol a modifier added,
// and a symbol it shares no colour with is one nothing may pay. So is
// every symbol when no colour was chosen.
func TestSpendOnlyLeavesNoColorUnpayable(t *testing.T) {
	c := spendOnlyCost(t, "{1}{B}{W/U}")
	c.SpendOnly = &ManaSpendOnly{Colors: []string{"W"}}
	got := c.foldSpendOnly(0)
	if want := [][]string{{noManaColor}, {"W"}, {"W"}}; !reflect.DeepEqual(requirementOptions(got), want) {
		t.Errorf("requirements %v, want %v", requirementOptions(got), want)
	}
	none := spendOnlyCost(t, "{3}")
	none.SpendOnly = (&ManaSpendOnly{ChosenColor: true}).ResolveFor(Card{})
	pool := ManaPool{{Color: "W"}, {Color: "U"}, {Color: "B"}, {Color: "R"}, {Color: "G"}, {Color: "C"}}
	if pool.CanPayFor(none.foldSpendOnly(0), 0, ManaSpendContext{Purpose: SpendPurposeActivate}) {
		t.Error("a chosen-colour cost with no colour chosen was paid")
	}
}

// ResolveFor reads the chosen colour off the source, keeps the declared
// colours, drops anything that is not a colour, and never aliases the
// declaration.
func TestSpendOnlyResolvesTheChosenColor(t *testing.T) {
	decl := &ManaSpendOnly{ChosenColor: true}
	got := decl.ResolveFor(Card{ChosenColor: "G"})
	if !reflect.DeepEqual(got.Colors, []string{"G"}) || got.ChosenColor {
		t.Errorf("resolved %+v, want Colors [G] and no ChosenColor", got)
	}
	if decl.Colors != nil {
		t.Error("resolving wrote into the declaration")
	}
	if got := (&ManaSpendOnly{Colors: []string{"b", "C", "B"}}).ResolveFor(Card{}); !reflect.DeepEqual(got.Colors, []string{"B"}) {
		t.Errorf("resolved %v, want [B]", got.Colors)
	}
	if (*ManaSpendOnly)(nil).ResolveFor(Card{ChosenColor: "W"}) != nil {
		t.Error("no clause resolved to one")
	}
}

// The pricer stamps the clause, resolved against the source; the price
// it renders is still the one printed.
func TestAbilityPricerStampsTheSpendOnlyClause(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := Card{InstanceID: uuid.New(), Name: "Throne", Controller: me.ID, Owner: me.ID, ChosenColor: "U"}
	ab := ActivatedAbilityShape{Cost: AbilityCost{Mana: "{3}", Tap: true, SpendOnly: &ManaSpendOnly{ChosenColor: true}}}
	var priced ParsedCost
	var err error
	g.ReadSnapshot(func() { priced, err = g.AbilityManaCostForEffect(me.ID, src, ZoneBattlefield, ab) })
	if err != nil {
		t.Fatal(err)
	}
	if priced.String() != "{3}" {
		t.Errorf("shown price %q, want the printed {3}", priced.String())
	}
	if priced.SpendOnly == nil || !reflect.DeepEqual(priced.SpendOnly.Colors, []string{"U"}) {
		t.Errorf("stamped %+v, want blue only", priced.SpendOnly)
	}
}

// The pool solver and the auto-tapper, both through costAsPaidByLocked:
// white pays, blue does not, and the planner taps only Plains.
func TestSpendOnlyPoolAndAutoTapAgree(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	cost := spendOnlyCost(t, "{3}")
	cost.SpendOnly = &ManaSpendOnly{Colors: []string{"W"}}
	ctx := ManaSpendContext{Purpose: SpendPurposeActivate}
	asPaid := func() ParsedCost {
		var out ParsedCost
		g.ReadSnapshot(func() { out = g.costAsPaidByLocked(me.ID, ctx, cost, 0) })
		return out
	}

	if (ManaPool{{Color: "W"}, {Color: "W"}, {Color: "U"}}).CanPayFor(asPaid(), 0, ctx) {
		t.Error("two white and a blue paid a white-only {3}")
	}
	if !(ManaPool{{Color: "W"}, {Color: "W"}, {Color: "W"}}).CanPayFor(asPaid(), 0, ctx) {
		t.Error("three white did not pay a white-only {3}")
	}

	plains := []uuid.UUID{
		pushBattlefieldForTest(g, me.ID, "Plains", "Basic Land — Plains", ""),
		pushBattlefieldForTest(g, me.ID, "Plains", "Basic Land — Plains", ""),
	}
	for i := 0; i < 3; i++ {
		pushBattlefieldForTest(g, me.ID, "Island", "Basic Land — Island", "")
	}
	if _, ok := g.AutoTapForCost(me.ID, asPaid(), 0); ok {
		t.Fatal("the auto-tapper planned a white-only {3} off two Plains and three Islands")
	}
	plains = append(plains, pushBattlefieldForTest(g, me.ID, "Plains", "Basic Land — Plains", ""))
	plan, ok := g.AutoTapForCost(me.ID, asPaid(), 0)
	if !ok {
		t.Fatal("no plan for a white-only {3} with three Plains")
	}
	if !sameIDSet(plan, plains) {
		t.Errorf("planned %v, want exactly the three Plains", plan)
	}
}

func sameIDSet(a, b []uuid.UUID) bool {
	if len(a) != len(b) {
		return false
	}
	seen := map[uuid.UUID]int{}
	for _, id := range a {
		seen[id]++
	}
	for _, id := range b {
		seen[id]--
	}
	for _, n := range seen {
		if n != 0 {
			return false
		}
	}
	return true
}

// CR 609.4b and the Celestial Dawn rulings: "spend only white mana" is
// a rule about how the cost may be paid, which is exactly what "spend
// mana as though it were mana of any color" changes — so under the
// grant any mana pays the folded symbols, real white first. A symbol
// with NO colour (nothing chosen) stays unpayable: there is no colour
// to spend anything as though it were.
func TestSpendOnlyUnderAnAnyColorGrant(t *testing.T) {
	g, me, _ := spendGrantTable(t)
	pushSpendGrant(g, me, anyColorYouOracle)
	ctx := ManaSpendContext{Purpose: SpendPurposeActivate}
	white := spendOnlyCost(t, "{3}")
	white.SpendOnly = &ManaSpendOnly{Colors: []string{"W"}}
	var asPaid ParsedCost
	g.ReadSnapshot(func() { asPaid = g.costAsPaidByLocked(me.ID, ctx, white, 0) })
	if !(ManaPool{{Color: "U"}, {Color: "C"}, {Color: "R"}}).CanPayFor(asPaid, 0, ctx) {
		t.Error("under the grant, blue, colorless and red did not pay a white-only {3}")
	}
	pool := ManaPool{{Color: "C"}, {Color: "W"}, {Color: "U"}, {Color: "W"}}
	spent, ok := pool.SpendManaFor(asPaid, 0, ctx)
	if !ok {
		t.Fatal("the payment failed")
	}
	whites := 0
	for _, tok := range spent {
		if tok.Color == "W" {
			whites++
		}
	}
	if whites != 2 {
		t.Errorf("spent %v, want both whites spent before anything else", spent)
	}

	unchosen := spendOnlyCost(t, "{3}")
	unchosen.SpendOnly = (&ManaSpendOnly{ChosenColor: true}).ResolveFor(Card{})
	g.ReadSnapshot(func() { asPaid = g.costAsPaidByLocked(me.ID, ctx, unchosen, 0) })
	if (ManaPool{{Color: "W"}, {Color: "U"}, {Color: "C"}}).CanPayFor(asPaid, 0, ctx) {
		t.Error("the grant paid a cost with no chosen colour")
	}
}

// The mana-side tags (CR 105.2a–b, CR 202.2d): monocolored is exactly
// one colour, multicolored two or more, a hybrid spell is both its
// colours, a colourless one is neither, and an object-less payment is
// neither either.
func TestMonocoloredAndMulticoloredTags(t *testing.T) {
	cast := func(colors ...string) ManaSpendContext {
		return ManaSpendContext{Purpose: SpendPurposeCast, Colors: colors}
	}
	cases := []struct {
		name       string
		ctx        ManaSpendContext
		mono, mult bool
	}{
		{"white spell", cast("W"), true, false},
		{"white-blue spell", cast("W", "U"), false, true},
		{"five-colour spell", cast("W", "U", "B", "R", "G"), false, true},
		{"colourless spell", cast(), false, false},
		{"repeated colour", cast("W", "w"), true, false},
		{"no object", ManaSpendContext{Colors: []string{"W"}}, false, false},
	}
	for _, tc := range cases {
		if got := tc.ctx.allows([]string{ManaRestrictMonocolored}); got != tc.mono {
			t.Errorf("%s: monocolored %v, want %v", tc.name, got, tc.mono)
		}
		if got := tc.ctx.allows([]string{ManaRestrictMulticolored}); got != tc.mult {
			t.Errorf("%s: multicolored %v, want %v", tc.name, got, tc.mult)
		}
	}
	// Throne's three tags together: a monocolored spell of THAT colour.
	throne := []string{ManaRestrictCast, ManaRestrictMonocolored, ManaRestrictColor("W")}
	if !cast("W").allows(throne) || cast("U").allows(throne) || cast("W", "U").allows(throne) {
		t.Error("Throne's mana: want a mono-white spell only")
	}
	if (ManaSpendContext{Purpose: SpendPurposeActivate, Colors: []string{"W"}}).allows(throne) {
		t.Error("Throne's mana paid an activation")
	}
}

// The tags ride the token, and the token rides the snapshot: a
// restricted pool comes back from JSON still restricted.
func TestMonocoloredTokenSurvivesASnapshotRoundTrip(t *testing.T) {
	g := newRestorableGame(t)
	me := g.Seats[0]
	me.ManaPool = ManaPool{{Color: "W", Restrictions: []string{ManaRestrictCast, ManaRestrictMonocolored, ManaRestrictColor("W")}}}
	_, restored := roundTrip(t, g)
	pool := restored.Seats[0].ManaPool
	if len(pool) != 1 {
		t.Fatalf("restored pool %v", pool)
	}
	w := spendOnlyCost(t, "{W}")
	if pool.CanPayFor(w, 0, ManaSpendContext{Purpose: SpendPurposeCast, Colors: []string{"W", "U"}}) {
		t.Error("the restored token paid for a multicolored spell")
	}
	if !pool.CanPayFor(w, 0, ManaSpendContext{Purpose: SpendPurposeCast, Colors: []string{"W"}}) {
		t.Error("the restored token refused a mono-white spell")
	}
}
