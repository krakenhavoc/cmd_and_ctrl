package game

import (
	"errors"
	"math"
	"runtime"
	"testing"

	"github.com/google/uuid"
)

// x_bound_test.go — #3895: every path that takes an X from an action
// refuses one above MaxAnnouncedX before any cost work, and an X at the
// bound with the mana to pay it still works.

// overBound is the announcements every path refuses: one past the
// bound, the largest int and a negative.
var overBound = []int{MaxAnnouncedX + 1, math.MaxInt, -1}

// maxRefusalBytes is the most a refused announcement may allocate. A
// refusal that reached the cost math at MaxAnnouncedX+1 would allocate
// tens of megabytes; one that stops at the gate allocates almost
// nothing.
const maxRefusalBytes = 1 << 20

// requireCheapRefusal runs `refuse` (which must refuse) and fails the
// test when it allocates more than maxRefusalBytes or more than a
// handful of times.
func requireCheapRefusal(t *testing.T, refuse func()) {
	t.Helper()
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	refuse()
	runtime.ReadMemStats(&after)
	if n := after.TotalAlloc - before.TotalAlloc; n > maxRefusalBytes {
		t.Errorf("the refusal allocated %d bytes, want at most %d", n, maxRefusalBytes)
	}
	if n := testing.AllocsPerRun(5, refuse); n > 64 {
		t.Errorf("the refusal made %.0f allocations, want at most 64", n)
	}
}

// --- casts ---------------------------------------------------------

const (
	xBoundSpendOnlyOracle = "test-x-bound-spend-only"
	xBoundPayLifeOracle   = "test-x-bound-pay-life"
	xBoundBlightOracle    = "test-x-bound-blight"
)

// withXBoundCatalog declares the cast-side X shapes the tests use:
// "Spend only black mana on X" (Drain Life), "pay X life" (Toxic
// Deluge) and "blight X" (Soul Immolation).
func withXBoundCatalog(t *testing.T) {
	t.Helper()
	prevSpend := CatalogSpellSpendOnly
	CatalogSpellSpendOnly = func(key string) *ManaSpendOnly {
		if key == xBoundSpendOnlyOracle {
			return &ManaSpendOnly{Colors: []string{"B"}, XOnly: true}
		}
		if prevSpend != nil {
			return prevSpend(key)
		}
		return nil
	}
	t.Cleanup(func() { CatalogSpellSpendOnly = prevSpend })
	prevAdd := CatalogAdditionalCost
	withCatalogAdditionalCost(t, func(key string) *AdditionalCost {
		switch key {
		case xBoundPayLifeOracle:
			return &AdditionalCost{PayLifeX: true, Label: "Pay X life"}
		case xBoundBlightOracle:
			return &AdditionalCost{BlightX: true, Label: "Blight X"}
		}
		if prevAdd != nil {
			return prevAdd(key)
		}
		return nil
	})
}

func xBoundFind(z *Zone, id uuid.UUID) *Card {
	for i := range z.Cards {
		if z.Cards[i].InstanceID == id {
			return &z.Cards[i]
		}
	}
	return nil
}

func xBoundSpell(p *Player, name, cost, oracle string) uuid.UUID {
	c := NewCard(name, p.ID)
	c.TypeLine = "Sorcery"
	c.ManaCost = cost
	c.Layout = "normal"
	c.OracleID = oracle
	p.Hand.PushTop(c)
	return c.InstanceID
}

// Every cast shape that reads X refuses one out of range, cheaply,
// with nothing moved or paid.
func TestCastRefusesAnXOutsideTheBound(t *testing.T) {
	withXBoundCatalog(t)
	for _, tc := range []struct{ name, cost, oracle string }{
		{"plain X", "{X}{R}", ""},
		{"spend only black on X", "{X}{B}", xBoundSpendOnlyOracle},
		{"pay X life", "{2}{B}{B}", xBoundPayLifeOracle},
		{"blight X", "{1}{B}", xBoundBlightOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			advanceTo(t, g, StepPrecombatMain)
			me := g.Seats[0]
			id := xBoundSpell(me, "Test "+tc.name, tc.cost, tc.oracle)
			life := me.Life
			for _, x := range overBound {
				for _, params := range []CastSpellParams{
					{XValue: x},
					{XValue: x, Strict: true, AutoTap: true},
					{XValue: x, ForceCast: true},
				} {
					if err := g.CastSpell(me.ID, id, params); !errors.Is(err, ErrInvalidParam) {
						t.Fatalf("X=%d %+v: err = %v, want ErrInvalidParam", x, params, err)
					}
				}
				requireCheapRefusal(t, func() { _ = g.CastSpell(me.ID, id, CastSpellParams{XValue: x}) })
			}
			if !me.Hand.Contains(id) || me.Life != life || len(g.StackMeta) != 0 {
				t.Fatalf("a refused cast moved something: in hand %v, life %d (was %d), %d stack items",
					me.Hand.Contains(id), me.Life, life, len(g.StackMeta))
			}
		})
	}
}

// The pricer the preview and the enumerator ask refuses the same X.
func TestPriceCastRefusesAnXOutsideTheBound(t *testing.T) {
	withXBoundCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	id := xBoundSpell(me, "Test Drain", "{X}{B}", xBoundSpendOnlyOracle)
	card := *xBoundFind(me.Hand, id)
	for _, x := range overBound {
		if _, err := g.PriceCast(me.ID, card, CastSpellParams{XValue: x}); !errors.Is(err, ErrInvalidParam) {
			t.Fatalf("X=%d: err = %v, want ErrInvalidParam", x, err)
		}
		requireCheapRefusal(t, func() { _, _ = g.PriceCast(me.ID, card, CastSpellParams{XValue: x}) })
	}
	if _, err := g.PriceCast(me.ID, card, CastSpellParams{XValue: MaxAnnouncedX}); err != nil {
		t.Fatalf("X at the bound: %v", err)
	}
}

// A real infinite-mana X: the bound itself, paid from a pool that
// holds it, casts.
func TestCastAtTheBoundWithTheManaStillCasts(t *testing.T) {
	withXBoundCatalog(t)
	for _, tc := range []struct {
		name, cost, oracle, color string
	}{
		{"plain X", "{X}{R}", "", "R"},
		{"spend only black on X", "{X}{B}", xBoundSpendOnlyOracle, "B"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			advanceTo(t, g, StepPrecombatMain)
			me := g.Seats[0]
			id := xBoundSpell(me, "Test "+tc.name, tc.cost, tc.oracle)
			me.ManaPool.AddMana(ManaToken{Color: tc.color, Count: MaxAnnouncedX + 1})
			if err := g.CastSpell(me.ID, id, CastSpellParams{XValue: MaxAnnouncedX, Strict: true}); err != nil {
				t.Fatalf("X=%d with %d mana: %v", MaxAnnouncedX, MaxAnnouncedX+1, err)
			}
			if got := g.StackMeta[id].XValue; got != MaxAnnouncedX {
				t.Errorf("XValue = %d, want %d", got, MaxAnnouncedX)
			}
			if n := me.ManaPool.Total(); n != 0 {
				t.Errorf("%d mana left in the pool, want 0", n)
			}
		})
	}
}

// The bound with only part of the mana is refused, and promptly: the
// pool solver stays linear in a folded X (each first-match walk resumes
// where the last stopped, and a failed augmenting path is not searched
// again), so this finishes in well under a second rather than minutes.
func TestSpendOnlyXAtTheBoundShortOfManaIsRefused(t *testing.T) {
	withXBoundCatalog(t)
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	id := xBoundSpell(me, "Test Drain", "{X}{B}", xBoundSpendOnlyOracle)
	const have = MaxAnnouncedX / 2
	me.ManaPool.AddMana(ManaToken{Color: "B", Count: have})
	err := g.CastSpell(me.ID, id, CastSpellParams{XValue: MaxAnnouncedX, Strict: true})
	var short *InsufficientManaError
	if !errors.As(err, &short) {
		t.Fatalf("err = %v, want *InsufficientManaError", err)
	}
	if want := MaxAnnouncedX + 1 - have; len(short.Missing) != want {
		t.Errorf("%d missing symbols, want %d", len(short.Missing), want)
	}
	if me.ManaPool.Total() != have || !me.Hand.Contains(id) {
		t.Error("a refused cast spent mana or left the hand")
	}
}

// A long generic shortfall is one {N}, so an X-sized shortfall is one
// symbol on the wire; a short one still lists each {1}.
func TestMissingNamesALongGenericShortfallOnce(t *testing.T) {
	cost, err := ParseCost("{X}{R}")
	if err != nil {
		t.Fatal(err)
	}
	pool := ManaPool{{Color: "R"}}
	if got := pool.Missing(cost, MaxAnnouncedX); len(got) != 1 || got[0] != "{1000000}" {
		t.Errorf("Missing at X=%d = %v, want [{1000000}]", MaxAnnouncedX, got)
	}
	if got := pool.Missing(cost, maxListedGeneric); len(got) != maxListedGeneric || got[0] != "{1}" {
		t.Errorf("Missing at X=%d = %v, want %d of {1}", maxListedGeneric, got, maxListedGeneric)
	}
}

// --- activations ---------------------------------------------------

// pushXBoundSource puts a permanent with one activated ability of
// `cost` under p, with `counters` on it.
func pushXBoundSource(g *Game, p *Player, typeLine string, cost AbilityCost, counters map[string]int) uuid.UUID {
	c := NewCard("X Bound Test", p.ID)
	c.TypeLine = typeLine
	c.Controller = p.ID
	c.Counters = counters
	c.ActivatedAbilities = []ActivatedAbilityShape{{
		Label:  "X bound test",
		Cost:   cost,
		Effect: func(*Game, *StackItem) error { return nil },
	}}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// Every activation shape that reads X — the mana half, a restricted
// X, loyalty −X, pay X {E}, discard X, tap X — refuses one out of
// range, cheaply, with nothing paid.
func TestActivationRefusesAnXOutsideTheBound(t *testing.T) {
	zero := 0
	for _, tc := range []struct {
		name     string
		typeLine string
		cost     AbilityCost
		counters map[string]int
	}{
		{"mana X", "Artifact", AbilityCost{Mana: "{X}"}, nil},
		{"two X", "Artifact", AbilityCost{Mana: "{X}{X}"}, nil},
		{"spend only black on X", "Artifact", AbilityCost{Mana: "{X}", SpendOnly: &ManaSpendOnly{Colors: []string{"B"}, XOnly: true}}, nil},
		{"loyalty X", "Legendary Planeswalker — Test", AbilityCost{Loyalty: &zero, LoyaltyX: true}, map[string]int{CounterLoyalty: 5}},
		{"energy X", "Artifact", AbilityCost{EnergyX: true}, nil},
		{"discard X", "Artifact", AbilityCost{DiscardCards: &DiscardCost{CountFromX: true}}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			advanceTo(t, g, StepPrecombatMain)
			me := g.Seats[0]
			src := pushXBoundSource(g, me, tc.typeLine, tc.cost, tc.counters)
			setEnergyForTest(t, g, me, 5)
			hand := len(me.Hand.Cards)
			for _, x := range overBound {
				if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{XValue: x}); !errors.Is(err, ErrInvalidParam) {
					t.Fatalf("X=%d: err = %v, want ErrInvalidParam", x, err)
				}
				requireCheapRefusal(t, func() { _ = g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{XValue: x}) })
			}
			if len(g.StackMeta) != 0 || PlayerEnergy(me) != 5 || len(me.Hand.Cards) != hand ||
				xBoundFind(g.Battlefield, src).Counters[CounterLoyalty] != tc.counters[CounterLoyalty] {
				t.Fatal("a refused activation paid something")
			}
		})
	}
}

// The bound itself activates when the cost can be paid: X mana from a
// pool that holds it, and −X loyalty from a source that has it.
func TestActivationAtTheBoundStillActivates(t *testing.T) {
	zero := 0
	for _, tc := range []struct {
		name     string
		typeLine string
		cost     AbilityCost
		counters map[string]int
		pool     int
	}{
		{"mana X", "Artifact", AbilityCost{Mana: "{X}"}, nil, MaxAnnouncedX},
		{"loyalty X", "Legendary Planeswalker — Test", AbilityCost{Loyalty: &zero, LoyaltyX: true}, map[string]int{CounterLoyalty: MaxAnnouncedX}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			advanceTo(t, g, StepPrecombatMain)
			me := g.Seats[0]
			src := pushXBoundSource(g, me, tc.typeLine, tc.cost, tc.counters)
			if tc.pool > 0 {
				me.ManaPool.AddMana(ManaToken{Color: "C", Count: tc.pool})
			}
			if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{XValue: MaxAnnouncedX}); err != nil {
				t.Fatalf("X=%d: %v", MaxAnnouncedX, err)
			}
			if len(g.StackMeta) != 1 {
				t.Fatalf("%d stack items, want 1", len(g.StackMeta))
			}
			for _, item := range g.StackMeta {
				if item.XValue != MaxAnnouncedX {
					t.Errorf("stack item X = %d, want %d", item.XValue, MaxAnnouncedX)
				}
			}
			if n := me.ManaPool.Total(); n != 0 {
				t.Errorf("%d mana left in the pool, want 0", n)
			}
		})
	}
}

// The sandbox verbs record an X on a stack item; they take the same
// range.
func TestSandboxAbilityVerbsRefuseAnXOutsideTheBound(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushBattlefieldForTest(g, me.ID, "Sandbox Source", "Artifact", "")
	for _, x := range overBound {
		if err := g.ActivateAbility(me.ID, src, AbilityParams{Label: "manual", XValue: x}); !errors.Is(err, ErrInvalidParam) {
			t.Fatalf("ActivateAbility X=%d: err = %v, want ErrInvalidParam", x, err)
		}
		if err := g.AnnounceTrigger(me.ID, src, AbilityParams{Label: "manual", XValue: x}); !errors.Is(err, ErrInvalidParam) {
			t.Fatalf("AnnounceTrigger X=%d: err = %v, want ErrInvalidParam", x, err)
		}
	}
	if len(g.StackMeta) != 0 || len(g.PendingTriggers) != 0 {
		t.Fatal("a refused announcement reached the stack")
	}
	if err := g.ActivateAbility(me.ID, src, AbilityParams{Label: "manual", XValue: MaxAnnouncedX}); err != nil {
		t.Fatalf("ActivateAbility at the bound: %v", err)
	}
	if err := g.AnnounceTrigger(me.ID, src, AbilityParams{Label: "manual", XValue: MaxAnnouncedX}); err != nil {
		t.Fatalf("AnnounceTrigger at the bound: %v", err)
	}
}

// --- the arithmetic ------------------------------------------------

// XMana and GenericWithX saturate rather than wrap, whatever X a
// caller hands them.
func TestXManaSaturatesPastTheBound(t *testing.T) {
	cost := ParsedCost{XSlots: 3, Generic: 5}
	if got := cost.XMana(MaxAnnouncedX); got != 3*MaxAnnouncedX {
		t.Errorf("XMana(bound) = %d, want %d", got, 3*MaxAnnouncedX)
	}
	if got := cost.GenericWithX(MaxAnnouncedX); got != 3*MaxAnnouncedX+5 {
		t.Errorf("GenericWithX(bound) = %d, want %d", got, 3*MaxAnnouncedX+5)
	}
	for _, x := range []int{math.MaxInt, math.MaxInt / 2, math.MaxInt/3 + 1} {
		if got := cost.XMana(x); got != maxManaDemand {
			t.Errorf("XMana(%d) = %d, want the saturation %d", x, got, maxManaDemand)
		}
		if got := cost.GenericWithX(x); got != maxManaDemand {
			t.Errorf("GenericWithX(%d) = %d, want the saturation %d", x, got, maxManaDemand)
		}
	}
	big := ParsedCost{XSlots: 1, Generic: math.MaxInt - 1}
	if got := big.GenericWithX(10); got != maxManaDemand {
		t.Errorf("GenericWithX on a huge generic = %d, want the saturation %d", got, maxManaDemand)
	}
}

// The spend-only fold never sizes an allocation by X: past what any
// announcement can reach it writes one requirement no mana can pay,
// and below it the requirements share one clipped options slice.
func TestFoldSpendOnlyIsBoundedByTheAnnouncement(t *testing.T) {
	cost, err := ParseCost("{X}{1}{B}")
	if err != nil {
		t.Fatal(err)
	}
	cost.SpendOnly = &ManaSpendOnly{Colors: []string{"B"}}

	var out ParsedCost
	requireCheapRefusal(t, func() { out = cost.foldSpendOnly(math.MaxInt) })
	if len(out.Required) != 2 || !unpayableByAnyMana(out.Required[1]) {
		t.Fatalf("fold at an unbounded X = %+v, want {B} and one requirement no mana pays", out.Required)
	}

	out = cost.foldSpendOnly(3)
	if len(out.Required) != 5 || out.Generic != 0 || out.XSlots != 0 {
		t.Fatalf("fold at X=3 = %+v, want {B} plus four black requirements", out)
	}
	grown := append(out.Required[1].Options, "W")
	if len(grown) != 2 || len(out.Required[2].Options) != 1 || out.Required[2].Options[0] != "B" {
		t.Fatalf("appending to one folded requirement's options changed another: %v", out.Required[2].Options)
	}
}
