package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// one_x_cost_test.go — #6085: one announced X pays an {X} mana symbol and
// a non-mana X in the same activation cost (CR 107.3a: "While an
// activated ability is on the stack, any X in its activation cost equals
// the announced value").

// talonCost is Talon of Pain's "{X}, {T}, Remove X charge counters from
// this artifact".
func talonCost() AbilityCost {
	return AbilityCost{Mana: "{X}", Tap: true, RemoveCounters: &CounterRemovalCost{
		Counter: CounterCharge, Variable: true, CountIsX: true,
	}}
}

// necronCost is Necron Overlord's "{X}, {T}, Tap X untapped artifacts you
// control".
func necronCost() AbilityCost {
	return AbilityCost{Mana: "{X}", Tap: true, TapOthers: tapOthersXArtifactCost()}
}

func oneXItem(t *testing.T, g *Game) *StackItem {
	t.Helper()
	if len(g.StackMeta) != 1 {
		t.Fatalf("%d stack items, want 1", len(g.StackMeta))
	}
	for _, item := range g.StackMeta {
		return item
	}
	return nil
}

// X=2 pays {2} AND removes two charge counters, with no count of its own
// on the wire: the announced X is the count.
func TestCountIsXRemovesTheAnnouncedX(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushXBoundSource(g, me, "Artifact", talonCost(), map[string]int{CounterCharge: 3})
	me.ManaPool.AddMana(ManaToken{Color: "C", Count: 2})

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{XValue: 2}); err != nil {
		t.Fatalf("activate at X=2: %v", err)
	}
	item := oneXItem(t, g)
	if item.XValue != 2 || item.Paid.CountersRemoved != 2 {
		t.Errorf("stack item X=%d, counters removed %d; want 2 and 2", item.XValue, item.Paid.CountersRemoved)
	}
	if n := counterOf(g, src, CounterCharge); n != 1 {
		t.Errorf("%d charge counters left, want 1", n)
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("%d mana left, want 0 — the {X} was not charged at X", n)
	}
}

// A count that agrees with X is accepted; one that disagrees is refused
// with nothing paid — one X cannot be two numbers.
func TestCountIsXRefusesADisagreeingCount(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushXBoundSource(g, me, "Artifact", talonCost(), map[string]int{CounterCharge: 3})
	me.ManaPool.AddMana(ManaToken{Color: "C", Count: 2})

	for _, counts := range [][]int{{1}, {3}, {1, 1}} {
		if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{XValue: 2, CounterCounts: counts}); !errors.Is(err, ErrInvalidParam) {
			t.Fatalf("X=2 with counts %v: err = %v, want ErrInvalidParam", counts, err)
		}
	}
	if counterOf(g, src, CounterCharge) != 3 || me.ManaPool.Total() != 2 || findBattlefieldCard(g, src).Tapped {
		t.Fatal("a refused activation paid part of its cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{XValue: 2, CounterCounts: []int{2}}); err != nil {
		t.Fatalf("X=2 with the agreeing count: %v", err)
	}
	if n := counterOf(g, src, CounterCharge); n != 1 {
		t.Errorf("%d charge counters left, want 1", n)
	}
}

// X above the counters held is refused (the counters bound X), with the
// pool, the source and its counters untouched.
func TestCountIsXRefusesMoreThanTheCountersHeld(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushXBoundSource(g, me, "Artifact", talonCost(), map[string]int{CounterCharge: 1})
	me.ManaPool.AddMana(ManaToken{Color: "C", Count: 2})

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{XValue: 2}); !errors.Is(err, ErrInsufficientCounters) {
		t.Fatalf("X=2 with one counter: err = %v, want ErrInsufficientCounters", err)
	}
	if counterOf(g, src, CounterCharge) != 1 || me.ManaPool.Total() != 2 || findBattlefieldCard(g, src).Tapped || len(g.StackMeta) != 0 {
		t.Fatal("a refused activation paid part of its cost")
	}
}

// X=0 is legal: nothing is removed and nothing is charged.
func TestCountIsXAtZero(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushXBoundSource(g, me, "Artifact", talonCost(), nil)

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate at X=0 with no counters: %v", err)
	}
	item := oneXItem(t, g)
	if item.XValue != 0 || item.Paid.CountersRemoved != 0 {
		t.Errorf("stack item X=%d, counters removed %d; want 0 and 0", item.XValue, item.Paid.CountersRemoved)
	}
	if !findBattlefieldCard(g, src).Tapped {
		t.Error("the {T} was not paid")
	}
}

// Necron Overlord's shape: X=2 pays {2} and taps two artifacts; a short
// tap payment, too few untapped artifacts, or too little mana for the X
// the taps announce are refused with nothing paid; X=0 is legal.
func TestManaXAndTapXShareOneX(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushXBoundSource(g, me, "Artifact", necronCost(), nil)
	one := pushTapOthersPermanent(g, me, "Rock One", "Artifact", false)
	two := pushTapOthersPermanent(g, me, "Rock Two", "Artifact", false)
	me.ManaPool.AddMana(ManaToken{Color: "C", Count: 2})

	untouched := func(what string) {
		t.Helper()
		if me.ManaPool.Total() != 2 || findBattlefieldCard(g, src).Tapped || findBattlefieldCard(g, one).Tapped ||
			findBattlefieldCard(g, two).Tapped || len(g.StackMeta) != 0 {
			t.Fatalf("%s: a refused activation paid part of its cost", what)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{XValue: 2, TapIDs: []uuid.UUID{one}}); !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("X=2 tapping one: err = %v, want ErrInvalidParam", err)
	}
	untouched("short tap")
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{XValue: 3, TapIDs: []uuid.UUID{one, two}}); err == nil {
		t.Fatal("X=3 with two untapped artifacts was accepted")
	}
	untouched("too few artifacts")

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{XValue: 2, TapIDs: []uuid.UUID{one, two}}); err != nil {
		t.Fatalf("activate at X=2: %v", err)
	}
	if item := oneXItem(t, g); item.XValue != 2 {
		t.Errorf("stack item X = %d, want 2", item.XValue)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left, want 0 — the {X} was not charged at X", me.ManaPool.Total())
	}
	if !findBattlefieldCard(g, one).Tapped || !findBattlefieldCard(g, two).Tapped || !findBattlefieldCard(g, src).Tapped {
		t.Error("the two artifacts and the source were not all tapped")
	}
}

func TestManaXAndTapXRefusesAnXThePoolCannotPay(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushXBoundSource(g, me, "Artifact", necronCost(), nil)
	one := pushTapOthersPermanent(g, me, "Rock One", "Artifact", false)
	two := pushTapOthersPermanent(g, me, "Rock Two", "Artifact", false)
	me.ManaPool.AddMana(ManaToken{Color: "C", Count: 1})

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{XValue: 2, TapIDs: []uuid.UUID{one, two}, Strict: true}); err == nil {
		t.Fatal("a strict X=2 with {1} in the pool was accepted")
	}
	if findBattlefieldCard(g, one).Tapped || findBattlefieldCard(g, two).Tapped || me.ManaPool.Total() != 1 {
		t.Fatal("a refused activation paid part of its cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate at X=0: %v", err)
	}
	if me.ManaPool.Total() != 1 || findBattlefieldCard(g, one).Tapped {
		t.Error("X=0 charged mana or tapped an artifact")
	}
}
