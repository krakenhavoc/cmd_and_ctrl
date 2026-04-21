package game

import (
	"testing"

	"github.com/google/uuid"
)

// mana_test.go covers the per-player ManaPool surface in isolation:
// AddMana / EmptyPool / CanPay / SpendMana / Missing. End-to-end
// activation flow (taps Sol Ring → pool fills → cast Lightning Bolt
// → strict-mode gate) lives in cards/effects/cards_test.go and
// mutations_test.go in later sub-PRs.

func costFor(t *testing.T, s string) ParsedCost {
	t.Helper()
	c, err := ParseCost(s)
	if err != nil {
		t.Fatalf("ParseCost(%q): %v", s, err)
	}
	return c
}

func TestManaPoolAddAndEmpty(t *testing.T) {
	var p ManaPool
	src := uuid.New()
	p.AddMana(ManaToken{Color: "G", Source: src}, ManaToken{Color: "C", Source: src})
	if got := len(p); got != 2 {
		t.Fatalf("len: got %d, want 2", got)
	}
	// Empty Color tokens dropped silently.
	p.AddMana(ManaToken{Color: "", Source: src})
	if got := len(p); got != 2 {
		t.Errorf("len after empty add: got %d, want 2", got)
	}
	cleared := p.EmptyPool()
	if cleared != 2 {
		t.Errorf("EmptyPool count: got %d, want 2", cleared)
	}
	if p != nil {
		t.Errorf("EmptyPool should leave nil slice, got %v", p)
	}
}

func TestManaPoolCanPaySimple(t *testing.T) {
	var p ManaPool
	p.AddMana(ManaToken{Color: "R"}, ManaToken{Color: "C"})
	// {1}{R} — generic 1 + one red.
	if !p.CanPay(costFor(t, "{1}{R}"), 0) {
		t.Errorf("expected pool to cover {1}{R}, got false")
	}
	// {2}{R} — short one generic.
	if p.CanPay(costFor(t, "{2}{R}"), 0) {
		t.Errorf("expected pool to fail {2}{R}")
	}
}

func TestManaPoolSpendDeductsAndPersists(t *testing.T) {
	var p ManaPool
	p.AddMana(
		ManaToken{Color: "R"},
		ManaToken{Color: "R"},
		ManaToken{Color: "C"},
		ManaToken{Color: "C"},
	)
	if ok := p.SpendMana(costFor(t, "{1}{R}"), 0); !ok {
		t.Fatalf("SpendMana failed")
	}
	// After {1}{R}: spent 1 red + 1 colorless. Pool should have
	// 1 red + 1 colorless left (preserving order).
	if got := len(p); got != 2 {
		t.Fatalf("post-spend len: got %d, want 2", got)
	}
	colors := []string{p[0].Color, p[1].Color}
	hasR := colors[0] == "R" || colors[1] == "R"
	hasC := colors[0] == "C" || colors[1] == "C"
	if !hasR || !hasC {
		t.Errorf("post-spend colors: got %v, want one R + one C", colors)
	}
}

func TestManaPoolSpendNoOpOnFailure(t *testing.T) {
	var p ManaPool
	p.AddMana(ManaToken{Color: "R"})
	before := len(p)
	if ok := p.SpendMana(costFor(t, "{2}{R}"), 0); ok {
		t.Fatalf("SpendMana succeeded on insufficient pool")
	}
	if got := len(p); got != before {
		t.Errorf("pool mutated despite failed spend: got %d, want %d", got, before)
	}
}

func TestManaPoolColorlessPreservedForLater(t *testing.T) {
	// Pool: 1R, 1C. Cost: {1}. Should spend the colorless first
	// and preserve the red for a future spell. This exercises the
	// "colorless first for generic" tiebreak.
	var p ManaPool
	p.AddMana(ManaToken{Color: "R"}, ManaToken{Color: "C"})
	if ok := p.SpendMana(costFor(t, "{1}"), 0); !ok {
		t.Fatalf("SpendMana failed")
	}
	if len(p) != 1 || p[0].Color != "R" {
		t.Errorf("expected R to survive after {1} spend, got %+v", p)
	}
}

func TestManaPoolHybridAcceptsEitherColor(t *testing.T) {
	// {W/U} should spend a single white OR a single blue.
	var p ManaPool
	p.AddMana(ManaToken{Color: "U"})
	if !p.CanPay(costFor(t, "{W/U}"), 0) {
		t.Errorf("hybrid {W/U} should accept blue")
	}
	p2 := ManaPool{}
	p2.AddMana(ManaToken{Color: "W"})
	if !p2.CanPay(costFor(t, "{W/U}"), 0) {
		t.Errorf("hybrid {W/U} should accept white")
	}
}

func TestManaPoolXSlotsScaleGeneric(t *testing.T) {
	// {X}{R} with X=3: needs 3 generic + 1 red.
	var p ManaPool
	for i := 0; i < 3; i++ {
		p.AddMana(ManaToken{Color: "C"})
	}
	p.AddMana(ManaToken{Color: "R"})
	if !p.CanPay(costFor(t, "{X}{R}"), 3) {
		t.Errorf("expected {X=3}{R} payable, got false")
	}
	if p.CanPay(costFor(t, "{X}{R}"), 4) {
		t.Errorf("expected {X=4}{R} unpayable on 3C+1R pool")
	}
}

func TestManaPoolMissingReportsShortfall(t *testing.T) {
	var p ManaPool
	p.AddMana(ManaToken{Color: "R"})
	missing := p.Missing(costFor(t, "{2}{R}{R}"), 0)
	// Pool has one R covering one R-requirement. The other R-requirement
	// + 2 generic are missing. Order: requirements first, then generic.
	if len(missing) != 3 {
		t.Fatalf("Missing: got %v, want 3 entries", missing)
	}
	if missing[0] != "{R}" {
		t.Errorf("Missing[0]: got %q, want {R}", missing[0])
	}
	if missing[1] != "{1}" || missing[2] != "{1}" {
		t.Errorf("Missing[1:3]: got %v, want two {1}", missing[1:])
	}
}

func TestManaPoolMissingNilWhenSatisfied(t *testing.T) {
	var p ManaPool
	p.AddMana(ManaToken{Color: "R"}, ManaToken{Color: "C"})
	if got := p.Missing(costFor(t, "{1}{R}"), 0); got != nil {
		t.Errorf("Missing on satisfied cost: got %v, want nil", got)
	}
}

func TestManaPoolEmptyCostNoOps(t *testing.T) {
	// A land's empty mana cost is payable trivially from any pool.
	var p ManaPool
	if !p.CanPay(costFor(t, ""), 0) {
		t.Errorf("empty cost should be payable from empty pool")
	}
}
