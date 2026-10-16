package game

import (
	"fmt"
	"math/rand/v2"
	"reflect"
	"slices"
	"testing"
)

// folded_run_test.go — #3909: the spend-only fold writes the restricted
// mana as one counted run, and every reader of a paid cost answers on
// the run exactly as it answered on the list the fold used to write.
// The randomized comparison is #3904's (coloured_pass_test.go), one
// step earlier: the same pools, checked against the unit fold.

// unitFoldSpendOnly is foldSpendOnly as it was before #3909, one
// requirement per restricted mana, kept as the reference the run fold
// is checked against.
func unitFoldSpendOnly(c ParsedCost, x int) ParsedCost {
	lim := c.SpendOnly
	if lim == nil {
		return c
	}
	if x < 0 {
		x = 0
	}
	allowed := lim.Colors
	if len(allowed) == 0 {
		allowed = []string{noManaColor}
	}
	out := c
	out.SpendOnly = nil
	n := c.XMana(x)
	out.XSlots = 0
	out.XReduced = 0
	req := make([]ColorRequirement, 0, len(c.Required))
	for _, r := range c.Required {
		if !lim.XOnly {
			r.Options = intersectColors(r.Options, allowed)
			if len(r.Options) == 0 {
				r.Options = []string{noManaColor}
			}
		}
		req = append(req, r)
	}
	if !lim.XOnly {
		n = satAddMana(n, c.Generic)
		out.Generic = 0
	}
	if n > maxFoldedMana {
		out.Required = append(req, ColorRequirement{Options: []string{noManaColor}})
		return out
	}
	opts := slices.Clip(append([]string(nil), allowed...))
	req = slices.Grow(req, n)
	for range n {
		req = append(req, ColorRequirement{Options: opts})
	}
	out.Required = req
	return out
}

// sameCost reports whether two costs, folded runs written out, list the
// same requirements in the same order and agree on every other field.
func sameCost(a, b ParsedCost) bool {
	a, b = a.unfolded(), b.unfolded()
	return a.Generic == b.Generic && a.XSlots == b.XSlots && a.XReduced == b.XReduced &&
		a.HasPhyrexian == b.HasPhyrexian && a.HasSnow == b.HasSnow &&
		reflect.DeepEqual(a.SpendOnly, b.SpendOnly) && sameRequirements(a.Required, b.Required) &&
		a.folded.n == 0 && b.folded.n == 0
}

// randomSpendOnlyCost is a random printed cost carrying a random
// "spend only" clause (whole cost or on X, any set of colours, none
// included), and an X to fold it at.
func randomSpendOnlyCost(r *rand.Rand) (ParsedCost, int) {
	c := ParsedCost{Generic: r.IntN(4), XSlots: r.IntN(3)}
	if c.XSlots > 0 && r.IntN(3) == 0 {
		c.XReduced = 1 + r.IntN(3)
	}
	for range r.IntN(4) {
		req := ColorRequirement{Options: passShapes[r.IntN(len(passShapes))]}
		req.Phyrexian = r.IntN(5) == 0
		req.LifeGranted = r.IntN(9) == 0
		c.HasPhyrexian = c.HasPhyrexian || req.Phyrexian
		c.Required = append(c.Required, req)
	}
	lim := &ManaSpendOnly{XOnly: r.IntN(2) == 0}
	for _, col := range passColors[:5] {
		if r.IntN(4) == 0 {
			lim.Colors = append(lim.Colors, col)
		}
	}
	c.SpendOnly = lim
	x := r.IntN(6)
	if r.IntN(6) == 0 {
		x = r.IntN(40)
	}
	return c, x
}

// asPaid puts the run fold and the unit fold through the same random
// steps a paid cost takes after the fold: a spend grant's widening (to
// any mana, or from some colours), a life-for-mana grant on the folded
// cost, and the Phyrexian strike.
func asPaid(t *testing.T, r *rand.Rand, run, unit ParsedCost, pool ManaPool) (ParsedCost, ParsedCost) {
	t.Helper()
	switch r.IntN(4) {
	case 1:
		anyType := r.IntN(2) == 0
		run, unit = widenForAnyColorSpend(run, anyType), widenForAnyColorSpend(unit, anyType)
	case 2:
		from := []string{passColors[r.IntN(5)]}
		run, unit = widenFromColors(run, from), widenFromColors(unit, from)
	case 3:
		colors := []string{passColors[r.IntN(5)]}
		run, unit = markLifeGranted(run, colors), markLifeGranted(unit, colors)
	}
	if n := r.IntN(3); n > 0 {
		var lifeRun, lifeUnit int
		run, lifeRun = PhyrexianLifePlan(run, pool, ManaSpendContext{}, n)
		unit, lifeUnit = PhyrexianLifePlan(unit, pool, ManaSpendContext{}, n)
		if lifeRun != lifeUnit {
			t.Fatalf("strike %d: run pays %d life, unit %d", n, lifeRun, lifeUnit)
		}
	}
	return run, unit
}

// withoutX is a spent record with the X marks (#3941) taken off and the
// runs they split joined again. Both sides of the comparison go through
// it, so it compares the mana and not how it was cut up: the unit fold,
// whose requirements are not marked, records no marks at all. The marks
// are checked on their own in mana_paid_x_test.go; this comparison is
// about WHICH mana paid.
func withoutX(spent []ManaToken) []ManaToken {
	var out []ManaToken
	for _, t := range spent {
		t.ForX = false
		out = appendRun(out, t)
	}
	return out
}

// slotsWithoutX is a picker's slots with the X marks taken off.
func slotsWithoutX(slots []PaperSlot) []PaperSlot {
	out := slices.Clone(slots)
	for i := range out {
		out[i].X = false
	}
	return out
}

// foldCheck is one reader's answer on the run and on the list.
type foldCheck struct {
	name      string
	got, want any
}

// The run fold is the unit fold: written out it is the same cost, and
// every reader of a paid cost gives the same answer on the run as on
// the list — the renderers and counts, and the pool solver's check,
// spend (both strategies), waived spend, missing list and paper
// shortfall — over random costs, clauses, X and pools.
func TestFoldedRunMatchesTheUnitFold(t *testing.T) {
	r := rand.New(rand.NewPCG(3909, 1))
	ctx := ManaSpendContext{}
	for c := range 4000 {
		printed, x := randomSpendOnlyCost(r)
		pool := randomPassPool(r)
		run, unit := asPaid(t, r, printed.foldSpendOnly(x), unitFoldSpendOnly(printed, x), pool)
		if !sameCost(run, unit) {
			t.Fatalf("case %d: %+v at X=%d: run %+v, unit %+v", c, printed, x, run, unit)
		}
		checks := []foldCheck{
			{"String", run.String(), unit.String()},
			{"Empty", run.Empty(), unit.Empty()},
			{"ManaValue", run.ManaValue(), unit.ManaValue()},
			{"totalManaWithX", run.totalManaWithX(x), unit.totalManaWithX(x)},
			{"PhyrexianSymbols", run.PhyrexianSymbols(), unit.PhyrexianSymbols()},
			{"printedPhyrexianSymbols", run.printedPhyrexianSymbols(), unit.printedPhyrexianSymbols()},
			{"LifeGrantedSymbols", run.LifeGrantedSymbols(), unit.LifeGrantedSymbols()},
			{"requiredCount", run.requiredCount(), len(unit.Required)},
			{"CanPayFor", pool.CanPayFor(run, x, ctx), pool.CanPayFor(unit, x, ctx)},
			{"MissingFor", pool.MissingFor(run, x, ctx), pool.MissingFor(unit, x, ctx)},
		}
		for _, strategy := range []ManaSpendStrategy{SpendPreserveColors, SpendDistinctColors} {
			pr, pu := slices.Clone(pool), slices.Clone(pool)
			sr, okr := pr.SpendManaForWith(run, x, ctx, strategy)
			su, oku := pu.SpendManaForWith(unit, x, ctx, strategy)
			wr, wu := slices.Clone(pool), slices.Clone(pool)
			wsr, fullr := wr.SpendWhatPaysForWith(run, x, ctx, strategy)
			wsu, fullu := wu.SpendWhatPaysForWith(unit, x, ctx, strategy)
			shr, shu := pool.shortfallFor(run, x, ctx, strategy), pool.shortfallFor(unit, x, ctx, strategy)
			s := fmt.Sprint(strategy)
			checks = append(checks,
				foldCheck{"spend " + s, fmt.Sprint(withoutX(sr), okr, pr), fmt.Sprint(withoutX(su), oku, pu)},
				foldCheck{"waived spend " + s, fmt.Sprint(withoutX(wsr), fullr, wr), fmt.Sprint(withoutX(wsu), fullu, wu)},
				foldCheck{"shortfall " + s, fmt.Sprint(shr.String(), slotsWithoutX(shr.slots())), fmt.Sprint(shu.String(), shu.slots())},
			)
		}
		for _, ch := range checks {
			if !reflect.DeepEqual(ch.got, ch.want) {
				t.Fatalf("case %d: %+v at X=%d, pool %v: %s on the run = %v, on the list = %v",
					c, printed, x, pool, ch.name, ch.got, ch.want)
			}
		}
	}
}

// The auto-tapper plans one requirement at a time and writes the run
// out first, so it plans, and tops up, exactly what it did for the list.
func TestAutoTapPlansTheFoldedRunAsTheList(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	for _, land := range []string{"Swamp", "Swamp", "Plains", "Island", "Forest", "Mountain"} {
		pushBattlefieldForTest(g, me.ID, land, "Basic Land — "+land, "")
	}
	r := rand.New(rand.NewPCG(3909, 2))
	for c := range 400 {
		printed, x := randomSpendOnlyCost(r)
		me.ManaPool = randomPassPool(r)
		run, unit := asPaid(t, r, printed.foldSpendOnly(x), unitFoldSpendOnly(printed, x), me.ManaPool)
		var (
			planRun, planUnit   tapPlan
			shortRun, shortUnit ParsedCost
			okRun, okUnit       bool
		)
		g.ReadSnapshot(func() {
			planRun, shortRun, okRun = g.autoTapTopUpLocked(me.ID, run, x, ManaSpendContext{}, nil, 0)
			planUnit, shortUnit, okUnit = g.autoTapTopUpLocked(me.ID, unit, x, ManaSpendContext{}, nil, 0)
		})
		if okRun != okUnit || !slices.Equal(planRun.cardIDs(), planUnit.cardIDs()) || !sameCost(shortRun, shortUnit) {
			t.Fatalf("case %d: %+v at X=%d, pool %v: top-up on the run %v %v %+v, on the list %v %v %+v",
				c, printed, x, me.ManaPool, okRun, planRun.cardIDs(), shortRun, okUnit, planUnit.cardIDs(), shortUnit)
		}
		idsRun, okRun := g.AutoTapForCost(me.ID, run, x)
		idsUnit, okUnit := g.AutoTapForCost(me.ID, unit, x)
		if okRun != okUnit || !slices.Equal(idsRun, idsUnit) {
			t.Fatalf("case %d: %+v at X=%d: plan on the run %v %v, on the list %v %v", c, printed, x, okRun, idsRun, okUnit, idsUnit)
		}
	}
}

// A "spend only black mana on X" payment at the X bound, fold included,
// is a handful of allocations whether it pays or comes up short.
func TestSpendOnlyXPaymentIsSmallAtTheBound(t *testing.T) {
	cost, err := ParseCost("{X}{B}")
	if err != nil {
		t.Fatal(err)
	}
	cost.SpendOnly = &ManaSpendOnly{Colors: []string{"B"}, XOnly: true}
	for _, black := range []int{MaxAnnouncedX + 1, MaxAnnouncedX / 2} {
		t.Run(fmt.Sprintf("black=%d", black), func(t *testing.T) {
			pool := ManaPool{ManaToken{Color: "B"}.withN(black)}
			requireSmall(t, 64<<10, func() {
				folded := cost.foldSpendOnly(MaxAnnouncedX)
				if pool.CanPayFor(folded, MaxAnnouncedX, ManaSpendContext{}) {
					p := slices.Clone(pool)
					if _, ok := p.SpendManaFor(folded, MaxAnnouncedX, ManaSpendContext{}); !ok || p.Total() != black-MaxAnnouncedX-1 {
						t.Errorf("spent to %v, want %d left", p, black-MaxAnnouncedX-1)
					}
				} else if got := pool.MissingFor(folded, MaxAnnouncedX, ManaSpendContext{}); len(got) != 1 {
					t.Errorf("missing %v, want one counted entry", got)
				}
			})
		})
	}
}
