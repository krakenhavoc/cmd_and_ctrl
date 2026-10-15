package game

import (
	"fmt"
	"math/rand/v2"
	"reflect"
	"runtime"
	"slices"
	"testing"
)

// coloured_pass_test.go — #3904: the coloured pass works run by run,
// books exactly what the unit-by-unit pass it replaced booked, and
// stays small at the X bound.

// unitColouredPass is the pass as it was before #3904, one unit per
// mana and one requirement at a time, kept as the reference the run
// pass is checked against.
func unitColouredPass(pool ManaPool, used []int, order []int, reqs []ColorRequirement) (deferred, unpaid []ColorRequirement) {
	if len(reqs) == 0 {
		return nil, nil
	}
	var units ManaPool
	var unitRun []int
	for _, i := range order {
		for range min(pool[i].N()-used[i], len(reqs)) {
			units = append(units, pool[i])
			unitRun = append(unitRun, i)
		}
	}
	unitOrder := make([]int, len(units))
	for u := range unitOrder {
		unitOrder[u] = u
	}
	booked := make([]bool, len(units))
	owner := make([]int, len(units))
	for i := range owner {
		owner[i] = -1
	}
	var prevOptions []string
	from, reassignFailed := 0, false
	for ri, req := range reqs {
		if ri == 0 || !slices.Equal(req.Options, prevOptions) {
			from, reassignFailed = 0, false
		}
		prevOptions = req.Options
		idx := unitFirstMatch(units, booked, unitOrder[from:], req)
		if idx >= 0 {
			from = idx + 1
		} else {
			from = len(units)
		}
		if idx < 0 && !req.AnyMana && !reassignFailed {
			idx = unitReassign(units, booked, unitOrder, owner, reqs, ri)
			reassignFailed = idx < 0
		}
		if idx < 0 {
			if req.AnyMana {
				deferred = append(deferred, req)
			} else {
				unpaid = append(unpaid, req)
			}
			continue
		}
		booked[idx] = true
		owner[idx] = ri
	}
	for u, b := range booked {
		if b {
			used[unitRun[u]]++
		}
	}
	return deferred, unpaid
}

func unitFirstMatch(work ManaPool, used []bool, order []int, req ColorRequirement) int {
	for _, i := range order {
		if !used[i] && matchColor(work[i].Color, req.Options) {
			return i
		}
	}
	return -1
}

func unitReassign(work ManaPool, used []bool, order []int, owner []int, reqs []ColorRequirement, ri int) int {
	visited := make([]bool, len(work))
	var augment func(r int) int
	augment = func(r int) int {
		for _, i := range order {
			if visited[i] || !matchColor(work[i].Color, reqs[r].Options) {
				continue
			}
			visited[i] = true
			if !used[i] {
				return i
			}
			prev := owner[i]
			if prev < 0 || reqs[prev].AnyMana {
				continue
			}
			if slices.Equal(reqs[prev].Options, reqs[r].Options) {
				continue
			}
			if j := augment(prev); j >= 0 {
				used[j] = true
				owner[j] = prev
				return i
			}
		}
		return -1
	}
	return augment(ri)
}

// expandReqRuns is runs one requirement at a time.
func expandReqRuns(runs []reqRun) []ColorRequirement {
	var out []ColorRequirement
	for _, r := range runs {
		for range r.n {
			out = append(out, r.req)
		}
	}
	return out
}

func sameRequirements(a, b []ColorRequirement) bool {
	return slices.EqualFunc(a, b, sameRequirement)
}

// passColors and passShapes are the mana and the requirement shapes the
// randomized comparisons draw from.
var (
	passColors = []string{"W", "U", "B", "R", "G", "C"}
	passShapes = [][]string{{"W"}, {"U"}, {"B"}, {"R"}, {"G"}, {"C"}, {"W", "U"}, {"B", "R"}, {"G", "W"}, {"U", "B"}, {"R", "G"}, {"B", "G", "U"}}
)

// randomPassPool is a random pool of up to six runs, mostly short and
// now and then long.
func randomPassPool(r *rand.Rand) ManaPool {
	var pool ManaPool
	for range r.IntN(7) {
		n := 1 + r.IntN(4)
		if r.IntN(8) == 0 {
			n = 1 + r.IntN(40)
		}
		pool = append(pool, ManaToken{Color: passColors[r.IntN(len(passColors))]}.withN(n))
	}
	return pool
}

// The run pass and the unit pass book the same mana and leave the same
// requirements unpaid and deferred, over random pools and costs built
// to exercise the walk, the augmenting path (hybrids that have to give
// a colour back), widened slots, used mana and long runs.
func TestColouredPassMatchesTheUnitPass(t *testing.T) {
	r := rand.New(rand.NewPCG(3904, 1))
	for c := range 4000 {
		pool := randomPassPool(r)
		used := make([]int, len(pool))
		for i := range pool {
			if r.IntN(4) == 0 {
				used[i] = r.IntN(pool[i].N() + 1)
			}
		}
		order := r.Perm(len(pool))
		order = order[:len(order)-r.IntN(min(2, len(order))+1)/2]
		var reqs []ColorRequirement
		for range r.IntN(8) {
			req := ColorRequirement{Options: passShapes[r.IntN(len(passShapes))]}
			req.Phyrexian = r.IntN(6) == 0
			req.AnyMana = r.IntN(7) == 0
			n := 1
			if r.IntN(3) == 0 {
				n = 1 + r.IntN(30)
			}
			for range n {
				reqs = append(reqs, req)
			}
		}

		wantUsed := slices.Clone(used)
		wantDeferred, wantUnpaid := unitColouredPass(pool, wantUsed, order, widenedLast(reqs))
		gotUsed := slices.Clone(used)
		deferred, unpaid := colouredPass(pool, gotUsed, order, requirementRuns(reqs, reqRun{}))
		if !slices.Equal(gotUsed, wantUsed) ||
			!sameRequirements(expandReqRuns(deferred), wantDeferred) ||
			!sameRequirements(expandReqRuns(unpaid), wantUnpaid) {
			t.Fatalf("case %d: pool %v used %v order %v reqs %v:\nruns used %v deferred %v unpaid %v\nunit used %v deferred %v unpaid %v",
				c, pool, used, order, reqs, gotUsed, expandReqRuns(deferred), expandReqRuns(unpaid), wantUsed, wantDeferred, wantUnpaid)
		}
	}
}

// The ADR 0118 §1 case the augmenting path exists for, by hand: {G/W}
// then a run of {G} out of [{G}, {W}, {G}] books all three, the hybrid
// moved onto the {W}.
func TestColouredPassMovesAHybridOffTheColourARunNeeds(t *testing.T) {
	pool := ManaPool{{Color: "G"}, {Color: "W"}, {Color: "G"}}
	used := make([]int, len(pool))
	reqs := []reqRun{
		{req: ColorRequirement{Options: []string{"G", "W"}}, n: 1},
		{req: ColorRequirement{Options: []string{"G"}}, n: 2},
	}
	_, unpaid := colouredPass(pool, used, []int{0, 1, 2}, reqs)
	if len(unpaid) != 0 || !slices.Equal(used, []int{1, 1, 1}) {
		t.Fatalf("used %v unpaid %v, want [1 1 1] and nothing unpaid", used, unpaid)
	}
}

// sameRequirement decides what a run is. A field it does not compare
// would merge two different requirements into one run.
func TestSameRequirementReadsEveryField(t *testing.T) {
	if n := reflect.TypeFor[ColorRequirement]().NumField(); n != 7 {
		t.Fatalf("ColorRequirement has %d fields; compare the new one in sameRequirement and update this count", n)
	}
	base := ColorRequirement{Options: []string{"B"}}
	for _, other := range []ColorRequirement{
		{Options: []string{"R"}},
		{Options: []string{"B"}, Phyrexian: true},
		{Options: []string{"B"}, NumericAlt: 2},
		{Options: []string{"B"}, HasNumericAlt: true},
		{Options: []string{"B"}, Snow: true},
		{Options: []string{"B"}, AnyMana: true},
		{Options: []string{"B"}, LifeGranted: true},
	} {
		if sameRequirement(base, other) {
			t.Errorf("sameRequirement(%+v, %+v) = true", base, other)
		}
	}
	if !sameRequirement(base, ColorRequirement{Options: []string{"B"}}) {
		t.Error("two plain {B} are not the same requirement")
	}
}

// requireSmall runs f and fails when it allocates more than maxBytes.
func requireSmall(t *testing.T, maxBytes uint64, f func()) {
	t.Helper()
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	f()
	runtime.ReadMemStats(&after)
	if n := after.TotalAlloc - before.TotalAlloc; n > maxBytes {
		t.Errorf("allocated %d bytes, want at most %d", n, maxBytes)
	}
}

// boundPassInput is a folded "spend only black mana on X" at the X
// bound behind a {B/R} hybrid, paid from a pool of `black` black mana
// and one red: the long run, the walk, and an augmenting path that has
// to move the hybrid onto the red.
func boundPassInput(black int) (ManaPool, []int, []reqRun) {
	pool := ManaPool{ManaToken{Color: "B"}.withN(black), {Color: "R"}}
	reqs := []reqRun{
		{req: ColorRequirement{Options: []string{"B", "R"}}, n: 1},
		{req: ColorRequirement{Options: []string{"B"}}, n: MaxAnnouncedX},
	}
	return pool, []int{0, 1}, reqs
}

// At the X bound the pass is a handful of allocations whether it pays
// or comes up short, and a short one names the shortfall as one run.
func TestColouredPassIsSmallAtTheBound(t *testing.T) {
	for _, tc := range []struct {
		name   string
		black  int
		unpaid int
	}{
		{"paid", MaxAnnouncedX, 0},
		{"half short", MaxAnnouncedX / 2, MaxAnnouncedX - MaxAnnouncedX/2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pool, order, reqs := boundPassInput(tc.black)
			run := func() {
				used := make([]int, len(pool))
				_, unpaid := colouredPass(pool, used, order, reqs)
				if got := reqRunsTotal(unpaid); got != tc.unpaid || len(unpaid) > 1 {
					t.Fatalf("unpaid %v, want one run of %d", unpaid, tc.unpaid)
				}
			}
			requireSmall(t, 64<<10, run)
			if n := testing.AllocsPerRun(5, run); n > 16 {
				t.Errorf("%.0f allocations, want at most 16", n)
			}
		})
	}
}

func BenchmarkColouredPassAtTheBound(b *testing.B) {
	for _, black := range []int{MaxAnnouncedX, MaxAnnouncedX / 2} {
		b.Run(fmt.Sprintf("black=%d", black), func(b *testing.B) {
			pool, order, reqs := boundPassInput(black)
			b.ReportAllocs()
			for range b.N {
				colouredPass(pool, make([]int, len(pool)), order, reqs)
			}
		})
	}
}

// The whole pool solver on a folded X at the bound, fold included:
// what a "spend only black mana on X" cast costs to check and to pay.
func BenchmarkSpendOnlyXPaymentAtTheBound(b *testing.B) {
	cost, err := ParseCost("{X}{B}")
	if err != nil {
		b.Fatal(err)
	}
	cost.SpendOnly = &ManaSpendOnly{Colors: []string{"B"}, XOnly: true}
	for _, black := range []int{MaxAnnouncedX + 1, MaxAnnouncedX / 2} {
		b.Run(fmt.Sprintf("black=%d", black), func(b *testing.B) {
			pool := ManaPool{ManaToken{Color: "B"}.withN(black)}
			b.ReportAllocs()
			for range b.N {
				folded := cost.foldSpendOnly(MaxAnnouncedX)
				if pool.CanPayFor(folded, MaxAnnouncedX, ManaSpendContext{}) {
					p := slices.Clone(pool)
					p.SpendManaFor(folded, MaxAnnouncedX, ManaSpendContext{})
				} else {
					pool.MissingFor(folded, MaxAnnouncedX, ManaSpendContext{})
				}
			}
		})
	}
}

// The same payment at an ordinary X, paid and one short.
func BenchmarkSpendOnlyXPayment(b *testing.B) {
	cost, err := ParseCost("{X}{B}")
	if err != nil {
		b.Fatal(err)
	}
	cost.SpendOnly = &ManaSpendOnly{Colors: []string{"B"}, XOnly: true}
	const x = 5
	for _, black := range []int{x + 1, x} {
		b.Run(fmt.Sprintf("x=%d/black=%d", x, black), func(b *testing.B) {
			pool := ManaPool{ManaToken{Color: "B"}.withN(black), {Color: "G"}}
			b.ReportAllocs()
			for range b.N {
				folded := cost.foldSpendOnly(x)
				if pool.CanPayFor(folded, x, ManaSpendContext{}) {
					p := slices.Clone(pool)
					p.SpendManaFor(folded, x, ManaSpendContext{})
				} else {
					pool.MissingFor(folded, x, ManaSpendContext{})
				}
			}
		})
	}
}
