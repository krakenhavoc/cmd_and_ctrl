package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// divide_test.go — #1563: the CR 601.2d gate and its default, as pure
// functions over an announcement's steps.

func dividedSteps(d DivideSpec, min, max int) []AnnouncedClause {
	c := TargetClause{Min: min, Max: max}
	c.Dividing(d)
	return []AnnouncedClause{{Mode: 0, Option: -1, Slot: 0, Clause: c}}
}

func refsOf(ids ...uuid.UUID) []TargetRef {
	out := make([]TargetRef, 0, len(ids))
	for _, id := range ids {
		out = append(out, TargetRef{Kind: TargetCard, ID: id})
	}
	return out
}

func TestDivideSpecTotalFor(t *testing.T) {
	cases := []struct {
		d    DivideSpec
		x    int
		want int
	}{
		{DivideSpec{Total: 4}, 9, 4},
		{DivideSpec{FromX: true}, 3, 3},
		{DivideSpec{FromX: true}, -2, 0},
		{DivideSpec{FromX: true, DoubleFromX: 6}, 5, 5},
		{DivideSpec{FromX: true, DoubleFromX: 6}, 6, 12},
		{DivideSpec{FromX: true, DoubleFromX: 6}, 7, 14},
	}
	for _, c := range cases {
		if got := c.d.TotalFor(c.x); got != c.want {
			t.Errorf("%+v at X=%d: %d, want %d", c.d, c.x, got, c.want)
		}
	}
	var nilSpec *DivideSpec
	if nilSpec.TotalFor(3) != 0 {
		t.Error("a nil spec divides nothing")
	}
}

func TestSettleDistribution(t *testing.T) {
	a, b, c := uuid.New(), uuid.New(), uuid.New()
	steps := dividedSteps(DivideSpec{Total: 4}, 0, 4)

	ok := []struct {
		name    string
		targets []TargetRef
		dist    map[uuid.UUID]int
		want    map[uuid.UUID]int
	}{
		{"chosen split", refsOf(a, b), map[uuid.UUID]int{a: 3, b: 1}, map[uuid.UUID]int{a: 3, b: 1}},
		{"lone target filled", refsOf(a), nil, map[uuid.UUID]int{a: 4}},
		{"lone target named", refsOf(a), map[uuid.UUID]int{a: 4}, map[uuid.UUID]int{a: 4}},
		{"no targets", nil, nil, nil},
		{"one each", refsOf(a, b, c), map[uuid.UUID]int{a: 2, b: 1, c: 1}, map[uuid.UUID]int{a: 2, b: 1, c: 1}},
	}
	for _, tc := range ok {
		got, err := settleDistribution(steps, tc.targets, tc.dist, 0)
		if err != nil {
			t.Errorf("%s: %v", tc.name, err)
			continue
		}
		if len(got) != len(tc.want) {
			t.Errorf("%s: %v, want %v", tc.name, got, tc.want)
		}
		for id, v := range tc.want {
			if got[id] != v {
				t.Errorf("%s: %v, want %v", tc.name, got, tc.want)
			}
		}
	}

	bad := []struct {
		name    string
		steps   []AnnouncedClause
		targets []TargetRef
		dist    map[uuid.UUID]int
	}{
		{"sum short", steps, refsOf(a, b), map[uuid.UUID]int{a: 2, b: 1}},
		{"sum over", steps, refsOf(a, b), map[uuid.UUID]int{a: 3, b: 2}},
		{"zero share", steps, refsOf(a, b), map[uuid.UUID]int{a: 4, b: 0}},
		{"missing share", steps, refsOf(a, b), map[uuid.UUID]int{a: 4}},
		{"share on a non-target", steps, refsOf(a, b), map[uuid.UUID]int{a: 2, b: 1, c: 1}},
		{"lone target, wrong amount", steps, refsOf(a), map[uuid.UUID]int{a: 1}},
		{"more targets than the amount", dividedSteps(DivideSpec{Total: 2}, 0, 0), refsOf(a, b, c), map[uuid.UUID]int{a: 1, b: 1, c: 0}},
		{"a division with nothing divided", []AnnouncedClause{{Option: -1, Clause: TargetClause{Min: 0, Max: 2}}}, refsOf(a), map[uuid.UUID]int{a: 1}},
	}
	for _, tc := range bad {
		if _, err := settleDistribution(tc.steps, tc.targets, tc.dist, 0); !errors.Is(err, ErrInvalidParam) {
			t.Errorf("%s: %v, want ErrInvalidParam", tc.name, err)
		}
	}

	// The free-form S13.1 path has no clause list and keeps what was sent.
	raw := map[uuid.UUID]int{a: 7}
	if got, err := settleDistribution(nil, refsOf(a), raw, 0); err != nil || got[a] != 7 {
		t.Errorf("free-form: %v %v, want the division unjudged", got, err)
	}
}

// #1657: "distribute UP TO that many" — the shares may fall short of
// the amount but not run over it, and each chosen target still gets 1.
func TestSettleDistributionUpTo(t *testing.T) {
	a, b, c := uuid.New(), uuid.New(), uuid.New()
	steps := dividedSteps(DivideSpec{Total: 4, UpTo: true}, 0, 0)
	for name, dist := range map[string]map[uuid.UUID]int{
		"short": {a: 1, b: 1},
		"exact": {a: 3, b: 1},
	} {
		if _, err := settleDistribution(steps, refsOf(a, b), dist, 0); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
	for name, dist := range map[string]map[uuid.UUID]int{
		"over":       {a: 3, b: 2},
		"zero share": {a: 1, b: 0},
	} {
		if _, err := settleDistribution(steps, refsOf(a, b), dist, 0); !errors.Is(err, ErrInvalidParam) {
			t.Errorf("%s: %v, want ErrInvalidParam", name, err)
		}
	}
	if _, err := settleDistribution(steps, refsOf(a, b, c), map[uuid.UUID]int{a: 1, b: 1, c: 1}, 0); err != nil {
		t.Errorf("three of up to four: %v", err)
	}
	if _, err := settleDistribution(dividedSteps(DivideSpec{Total: 2, UpTo: true}, 0, 0), refsOf(a, b, c),
		map[uuid.UUID]int{a: 1, b: 1, c: 1}, 0); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("three targets cannot share up to 2 at 1 each: %v", err)
	}
	// The even split is still the whole amount, which "up to" allows.
	got, ok := EvenDistribution(steps, refsOf(a, b), 0)
	if !ok || got[a] != 2 || got[b] != 2 {
		t.Errorf("even split of up to 4: %v %v", got, ok)
	}
}

// #1657: an amount rule is read once, when the steps are bound, and the
// bound step carries a plain amount — the catalog's clause untouched.
var testDivideAmountSeen DivideAmountArgs

var testDivideAmount = RegisterDivideAmount("test-divide-amount", func(_ *Game, a DivideAmountArgs) DivideSpec {
	testDivideAmountSeen = a
	if a.AltCost == "madness" {
		return DivideSpec{FromX: true}
	}
	return DivideSpec{Total: 3}
})

func TestBindDivideAmounts(t *testing.T) {
	g := &Game{}
	a, b := uuid.New(), uuid.New()
	clause := TargetClause{Min: 0, Max: 0}
	clause.Dividing(DivideSpec{AmountKey: testDivideAmount, UpTo: true})
	shared := clause.Divide

	steps := []AnnouncedClause{{Option: -1, Clause: clause}}
	who, src := uuid.New(), uuid.New()
	g.bindDivideAmountsLocked(steps, DivideAmountArgs{Controller: who, Source: src})
	d := steps[0].Clause.Divide
	if d == shared || !d.AmountKey.IsZero() || d.Total != 3 || !d.UpTo {
		t.Fatalf("bound divide = %+v, want a fresh Total 3 that keeps UpTo", d)
	}
	if shared.AmountKey.IsZero() || shared.Total != 0 {
		t.Error("binding wrote through to the catalog's clause")
	}
	if testDivideAmountSeen.Controller != who || testDivideAmountSeen.Source != src {
		t.Errorf("the rule saw %+v", testDivideAmountSeen)
	}
	if _, err := settleDistribution(steps, refsOf(a, b), map[uuid.UUID]int{a: 1, b: 1}, 0); err != nil {
		t.Errorf("up to 3: %v", err)
	}

	// The claimed alternative cost reaches the rule, and an X answer is
	// sized by the announced X.
	steps = []AnnouncedClause{{Option: -1, Clause: clause}}
	g.bindDivideAmountsLocked(steps, DivideAmountArgs{AltCost: "madness"})
	if got := steps[0].Clause.Divide.TotalFor(5); got != 5 {
		t.Errorf("madness answer at X=5: %d", got)
	}

	// An unregistered key sizes the division at 0 — nothing can be
	// chosen — rather than guessing.
	unknown := &DivideSpec{AmountKey: DivideAmount{key: "no-such-rule"}}
	if got := g.divideAmountLocked(unknown, DivideAmountArgs{}); got.TotalFor(9) != 0 {
		t.Errorf("unknown rule: %+v", got)
	}
}

func TestEvenDistribution(t *testing.T) {
	a, b, c := uuid.New(), uuid.New(), uuid.New()
	steps := dividedSteps(DivideSpec{FromX: true}, 0, 0)
	got, ok := EvenDistribution(steps, refsOf(a, b, c), 5)
	if !ok || got[a] != 2 || got[b] != 2 || got[c] != 1 {
		t.Errorf("5 over three: %v %v, want 2/2/1", got, ok)
	}
	if _, err := settleDistribution(steps, refsOf(a, b, c), got, 5); err != nil {
		t.Errorf("the default must pass the gate: %v", err)
	}
	if _, ok := EvenDistribution(steps, refsOf(a, b, c), 2); ok {
		t.Error("three targets cannot divide 2")
	}
	if got, ok := EvenDistribution(steps, nil, 5); !ok || got != nil {
		t.Errorf("no targets: %v %v", got, ok)
	}
}
