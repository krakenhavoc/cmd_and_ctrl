package game

import (
	"encoding/json"
	"math/rand/v2"
	"slices"
	"testing"
)

// mana_paid_x_test.go — #3941 (ADR 0068, amendment 2026-10-16): the pool
// solver records which spent mana paid the restricted {X}, as ForX on
// the run, and a spell reads it through ManaSpent.OnX.

// onXCost is {X}{2}{B} with "spend only black and/or red mana on X",
// folded for x: the cost Soul Burn pays.
func onXCost(t *testing.T, x int, colors ...string) ParsedCost {
	t.Helper()
	cost, err := ParseCost("{X}{2}{B}")
	if err != nil {
		t.Fatal(err)
	}
	cost.SpendOnly = &ManaSpendOnly{Colors: colors, XOnly: true}
	return cost.foldSpendOnly(x)
}

func poolOf(spec string) ManaPool {
	var p ManaPool
	for _, r := range spec {
		p = append(p, ManaToken{Color: string(r)})
	}
	return p
}

// spendFrom spends from a pool the caller does not need back.
func spendFrom(pool ManaPool, cost ParsedCost, x int, ctx ManaSpendContext) ([]ManaToken, bool) {
	return pool.SpendManaFor(cost, x, ctx)
}

func forX(spent []ManaToken, color string) int {
	return ManaSpent{tokens: spent}.OnX().Count(color)
}

// The X part is the folded run and nothing else: the printed {B} and the
// generic {2} are never marked, and the marked mana is exactly X.
func TestSpentManaIsMarkedForTheXPart(t *testing.T) {
	spent, ok := spendFrom(poolOf("RRBBB"), onXCost(t, 2, "B", "R"), 2, ManaSpendContext{})
	if !ok {
		t.Fatal("payable pool refused")
	}
	if got := (ManaSpent{tokens: spent}).Total(); got != 5 {
		t.Errorf("%d spent, want 5", got)
	}
	if got := (ManaSpent{tokens: spent}).OnX().Total(); got != 2 {
		t.Errorf("%d marked for X, want exactly X = 2: %+v", got, spent)
	}
}

// Drain Life's X is black only, and the printed {B} is black too. The two
// requirements are identical, and the X part must still be its own run:
// X black on X, one more black on the {B}.
func TestTheXPartIsNotMergedWithAnIdenticalPrintedSymbol(t *testing.T) {
	cost, err := ParseCost("{X}{1}{B}")
	if err != nil {
		t.Fatal(err)
	}
	cost.SpendOnly = &ManaSpendOnly{Colors: []string{"B"}, XOnly: true}
	spent, ok := spendFrom(poolOf("BBBBB"), cost.foldSpendOnly(3), 3, ManaSpendContext{})
	if !ok {
		t.Fatal("refused")
	}
	if got := forX(spent, "B"); got != 3 {
		t.Errorf("%d black marked for X, want 3: %+v", got, spent)
	}
}

// An unrestricted X is generic mana, which any mana pays: it is not
// recorded, and neither is the whole-cost form of the clause.
func TestOnlyARestrictedXIsRecorded(t *testing.T) {
	plain, err := ParseCost("{X}{B}")
	if err != nil {
		t.Fatal(err)
	}
	spent, ok := spendFrom(poolOf("BBB"), plain, 2, ManaSpendContext{})
	if !ok || forX(spent, "B") != 0 {
		t.Errorf("an unrestricted X was recorded: %+v", spent)
	}
	whole, err := ParseCost("{3}")
	if err != nil {
		t.Fatal(err)
	}
	whole.SpendOnly = &ManaSpendOnly{Colors: []string{"W"}}
	spent, ok = spendFrom(poolOf("WWW"), whole.foldSpendOnly(0), 0, ManaSpendContext{})
	if !ok || forX(spent, "W") != 0 {
		t.Errorf("a whole-cost restriction was recorded as X: %+v", spent)
	}
}

// The wish decides which colour the X part takes when it has a choice
// and nothing else: red sits ahead of black in the pool, and without the
// wish the walk spends the red first.
func TestTheXPartTakesTheFavouredColourFirst(t *testing.T) {
	pool := poolOf("RRBBB")
	cost := onXCost(t, 2, "B", "R")
	plain, _ := spendFrom(slices.Clone(pool), cost, 2, ManaSpendContext{})
	if got := forX(plain, "B"); got != 0 {
		t.Fatalf("without a wish %d black went on X, want the walk's red first (0)", got)
	}
	wished := ManaSpendContext{wish: &spendWish{colors: []string{"B"}}}
	spent, _ := spendFrom(slices.Clone(pool), cost, 2, wished)
	if got := forX(spent, "B"); got != 2 {
		t.Errorf("with the wish %d black went on X, want 2: %+v", got, spent)
	}
	// The printed {B} is paid first, so the wish cannot take its black.
	spent, _ = spendFrom(poolOf("BRRRR"), cost, 2, wished)
	if got := forX(spent, "B"); got != 0 {
		t.Errorf("the only black paid the printed {B}, but %d is on X", got)
	}
}

// The wish never changes whether a cost is payable, how much it spends or
// how much of it is X: it picks WHICH of the legal mana pays X.
func TestTheXWishNeverChangesWhatIsPayable(t *testing.T) {
	r := rand.New(rand.NewPCG(3941, 1))
	wished := ManaSpendContext{wish: &spendWish{colors: []string{"B", "R"}}}
	plain := ManaSpendContext{}
	for c := range 3000 {
		printed, x := randomSpendOnlyCost(r)
		printed.SpendOnly.XOnly = true
		cost := printed.foldSpendOnly(x)
		pool := randomPassPool(r)
		if a, b := pool.CanPayFor(cost, x, plain), pool.CanPayFor(cost, x, wished); a != b {
			t.Fatalf("case %d: %+v at X=%d, pool %v: payable %v plain, %v wished", c, printed, x, pool, a, b)
		}
		pa, pb := slices.Clone(pool), slices.Clone(pool)
		sa, oka := pa.SpendManaFor(cost, x, plain)
		sb, okb := pb.SpendManaFor(cost, x, wished)
		if oka != okb {
			t.Fatalf("case %d: spend ok %v plain, %v wished", c, oka, okb)
		}
		if !oka {
			continue
		}
		ta, tb := (ManaSpent{tokens: sa}), (ManaSpent{tokens: sb})
		if ta.Total() != tb.Total() || ta.OnX().Total() != tb.OnX().Total() || TotalMana(pa) != TotalMana(pb) {
			t.Fatalf("case %d: %+v at X=%d, pool %v: spent %d (%d on X) plain, %d (%d on X) wished",
				c, printed, x, pool, ta.Total(), ta.OnX().Total(), tb.Total(), tb.OnX().Total())
		}
		if got, want := ta.OnX().Total(), cost.folded.n; got != want {
			t.Fatalf("case %d: %d marked for X, want the run's %d", c, got, want)
		}
		// Favouring black and red cannot put fewer of them on X.
		if a, b := ta.OnX().Count("B")+ta.OnX().Count("R"), tb.OnX().Count("B")+tb.OnX().Count("R"); b < a {
			t.Fatalf("case %d: favoured colours on X fell from %d to %d", c, a, b)
		}
	}
}

// Marked runs are different mana from unmarked ones, so a run never
// absorbs the other.
func TestAMarkedRunDoesNotMergeWithAnUnmarkedOne(t *testing.T) {
	a := ManaToken{Color: "B"}
	b := ManaToken{Color: "B", ForX: true}
	if a.sameRun(b) || b.sameRun(a) {
		t.Fatal("a marked run joined an unmarked one")
	}
	runs := appendRun(appendRun(nil, a), b)
	if len(runs) != 2 {
		t.Errorf("%d runs, want 2: %+v", len(runs), runs)
	}
}

// The mark is part of the record and survives a snapshot round trip.
func TestTheXMarkSurvivesTheWire(t *testing.T) {
	in := PaidCost{Mana: []ManaToken{{Color: "B", ForX: true, Count: 3}, {Color: "R"}}}
	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out PaidCost
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	if got := out.Spent().OnX().Count("B"); got != 3 || out.Spent().Count("R") != 1 {
		t.Errorf("round trip lost the mark: %s -> %+v", raw, out.Mana)
	}
}

// --- a declared paper payment ------------------------------------------

// shortByOneXAndTwo is the shortfall of Soul Burn at X = 2 off two black.
func shortByOneXAndTwo(t *testing.T, wish bool) paperShortfall {
	t.Helper()
	ctx := ManaSpendContext{}
	if wish {
		ctx.wish = &spendWish{colors: []string{"B"}}
	}
	s := poolOf("BB").shortfallFor(onXCost(t, 2, "B", "R"), 2, ctx, SpendPreserveColors)
	if s.amount() != 3 {
		t.Fatalf("shortfall = %s, want three", s.String())
	}
	return s
}

func TestAMarkedDeclarationRecordsTheCastersAllocation(t *testing.T) {
	s := shortByOneXAndTwo(t, false)
	slots := s.slots()
	if len(slots) != 3 || !slots[0].X || slots[1].X || slots[2].X {
		t.Fatalf("slots = %+v, want the first marked X", slots)
	}
	toks, err := s.tokensFor(PaperPayment{Mana: []PaperMana{{Color: "R", X: true}, {Color: "B"}, {Color: "C"}}})
	if err != nil {
		t.Fatal(err)
	}
	onX := ManaSpent{tokens: toks}.OnX()
	if onX.Total() != 1 || onX.Count("R") != 1 {
		t.Errorf("on X = %+v, want the one red the caster marked", toks)
	}
}

func TestAnUnmarkedDeclarationGivesXTheFavouredColour(t *testing.T) {
	declared := PaperPayment{Mana: []PaperMana{{Color: "R"}, {Color: "B"}, {Color: "C"}}}
	toks, err := shortByOneXAndTwo(t, true).tokensFor(declared)
	if err != nil {
		t.Fatal(err)
	}
	if got := forX(toks, "B"); got != 1 {
		t.Errorf("favoured: %d black on X, want 1: %+v", got, toks)
	}
	// Total declared mana is all in the record, marked or not.
	if got := (ManaSpent{tokens: toks}).Total(); got != 3 {
		t.Errorf("%d recorded, want the 3 declared", got)
	}
	toks, err = shortByOneXAndTwo(t, false).tokensFor(declared)
	if err != nil {
		t.Fatal(err)
	}
	if got := forX(toks, "B"); got != 0 {
		t.Errorf("unfavoured walk: %d black on X, want the red declared first (0)", got)
	}
}

func TestABadXMarkIsRefused(t *testing.T) {
	s := shortByOneXAndTwo(t, false)
	for name, mana := range map[string][]PaperMana{
		"a colour X does not admit": {{Color: "G", X: true}, {Color: "B"}, {Color: "C"}},
		"more marked than X slots":  {{Color: "B", X: true}, {Color: "B", X: true}, {Color: "C"}},
		"the generic marked":        {{Color: "B"}, {Color: "C", X: true}, {Color: "C"}},
	} {
		if _, err := s.tokensFor(PaperPayment{Mana: mana}); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	// A mark on a shortfall with no X slot is refused.
	plain, _ := ParseCost("{1}{B}")
	short := ManaPool(nil).shortfallFor(plain, 0, ManaSpendContext{}, SpendPreserveColors)
	if _, err := short.tokensFor(PaperPayment{Mana: []PaperMana{{Color: "B"}, {Color: "C", X: true}}}); err == nil {
		t.Error("a mark with no X slot was accepted")
	}
}

// A shortfall the size of an announced X is one counted slot, answered by
// one counted entry, and the mark rides on it.
func TestACountedXSlotTakesACountedMark(t *testing.T) {
	cost := onXCost(t, MaxAnnouncedX, "B")
	short := ManaPool(nil).shortfallFor(cost, MaxAnnouncedX, ManaSpendContext{}, SpendPreserveColors)
	var xSlot *PaperSlot
	for _, s := range short.slots() {
		if s.X {
			xSlot = &s
		}
	}
	if xSlot == nil || xSlot.Count != MaxAnnouncedX {
		t.Fatalf("slots = %+v, want one counted X slot of %d", short.slots(), MaxAnnouncedX)
	}
	toks, err := short.tokensFor(PaperPayment{Mana: []PaperMana{
		{Color: "B", Count: MaxAnnouncedX, X: true}, {Color: "C"}, {Color: "C"}, {Color: "B"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if got := forX(toks, "B"); got != MaxAnnouncedX {
		t.Errorf("%d marked, want %d", got, MaxAnnouncedX)
	}
}
