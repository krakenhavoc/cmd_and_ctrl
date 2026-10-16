package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// counter_cost_self_any_kinds_test.go — #3987: "Remove two counters from
// Ghost-Spider". The any-kind among payment of #943 with the SOURCE as its
// one candidate: the counters may be of any kinds the source holds, the
// activator names the kinds and how many of each (CR 602.2b, 118.3), and
// nothing but the source pays.

// selfAnyKindsCost is the shape under test: two counters of any kinds off
// the source.
func selfAnyKindsCost() AbilityCost {
	return AbilityCost{RemoveCounters: &CounterRemovalCost{N: 2, Among: true}}
}

// pushSelfAnyKindsSource seats the source with the given counters.
func pushSelfAnyKindsSource(g *Game, owner *Player, counters map[string]int) uuid.UUID {
	src := pushCounterCostSource(g, owner, selfAnyKindsCost(), nil)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == src {
			g.Battlefield.Cards[i].Counters = counters
		}
	}
	return src
}

// The headline: a +1/+1 and a shield counter off the source is one legal
// payment, sent the way the client sends it (no permanents, one count and
// one kind per part).
func TestSelfAnyKindsCounterCostMixesKinds(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushSelfAnyKindsSource(g, me, map[string]int{CounterPlusOne: 1, CounterShield: 1})

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{
		CounterCounts: []int{1, 1},
		CounterKinds:  []string{CounterPlusOne, CounterShield},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if counterOf(g, src, CounterPlusOne) != 0 || counterOf(g, src, CounterShield) != 0 {
		t.Errorf("counters left: %+v", g.Battlefield.Cards)
	}
}

// One kind is the #625 wire shape: a single counter_kind and either no
// count (the printed two) or the count spelled out, with or without the
// source named.
func TestSelfAnyKindsCounterCostAcceptsOneKind(t *testing.T) {
	for _, tc := range []struct {
		name string
		pay  func(src uuid.UUID) ActivateAbilityParams
	}{
		{"no count", func(uuid.UUID) ActivateAbilityParams {
			return ActivateAbilityParams{CounterKind: CounterMinusOne}
		}},
		{"the count", func(uuid.UUID) ActivateAbilityParams {
			return ActivateAbilityParams{CounterKind: CounterMinusOne, CounterCounts: []int{2}}
		}},
		{"the source named", func(src uuid.UUID) ActivateAbilityParams {
			return ActivateAbilityParams{CounterSourceIDs: []uuid.UUID{src}, CounterKind: CounterMinusOne, CounterCounts: []int{2}}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			advanceTo(t, g, StepPrecombatMain)
			me := g.Seats[0]
			src := pushSelfAnyKindsSource(g, me, map[string]int{CounterMinusOne: 3})
			if err := g.ActivateCatalogAbility(me.ID, src, 0, tc.pay(src)); err != nil {
				t.Fatalf("activate: %v", err)
			}
			if got := counterOf(g, src, CounterMinusOne); got != 1 {
				t.Errorf("source has %d -1/-1 counters, want 1", got)
			}
		})
	}
}

// Every refusal, and a refused payment removes nothing.
func TestSelfAnyKindsCounterCostIsValidatedAsASet(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushSelfAnyKindsSource(g, me, map[string]int{CounterPlusOne: 2, CounterShield: 1})
	bear := pushPermanentWithCounters(g, me, "Bear", "Creature — Bear", map[string]int{CounterPlusOne: 2})

	cases := []struct {
		name   string
		ids    []uuid.UUID
		counts []int
		kinds  []string
		kind   string
		want   error
	}{
		{"another permanent's counters", []uuid.UUID{bear}, []int{2}, nil, CounterPlusOne, ErrInvalidParam},
		{"the source and another permanent", []uuid.UUID{src, bear}, []int{1, 1}, []string{CounterPlusOne, CounterPlusOne}, "", ErrInvalidParam},
		{"more of a kind than the source holds", nil, []int{2}, nil, CounterShield, ErrInsufficientCounters},
		{"the same kind twice", nil, []int{1, 1}, []string{CounterPlusOne, CounterPlusOne}, "", ErrInvalidParam},
		{"short of two", nil, []int{1}, nil, CounterShield, ErrInvalidParam},
		{"over two", nil, []int{2, 1}, []string{CounterPlusOne, CounterShield}, "", ErrInvalidParam},
		{"a kind the source does not hold", nil, []int{1, 1}, []string{CounterPlusOne, CounterStun}, "", ErrInsufficientCounters},
		{"no kind named", nil, []int{1, 1}, nil, "", ErrInvalidParam},
		{"a part that removes nothing", nil, []int{2, 0}, []string{CounterPlusOne, CounterShield}, "", ErrInvalidParam},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{
				CounterSourceIDs: tc.ids,
				CounterCounts:    tc.counts,
				CounterKinds:     tc.kinds,
				CounterKind:      tc.kind,
			})
			if !errors.Is(err, tc.want) {
				t.Fatalf("activate: %v, want %v", err, tc.want)
			}
			if counterOf(g, src, CounterPlusOne) != 2 || counterOf(g, src, CounterShield) != 1 ||
				counterOf(g, bear, CounterPlusOne) != 2 {
				t.Fatal("a refused payment removed counters")
			}
		})
	}
}

// The candidate walk offers only the source, every kind it holds, and the
// cost is payable exactly when those kinds total two.
func TestSelfAnyKindsCounterCostOptionsAndPayable(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	rc := selfAnyKindsCost().RemoveCounters
	src := pushSelfAnyKindsSource(g, me, map[string]int{CounterPlusOne: 1})
	pushPermanentWithCounters(g, me, "Bear", "Creature — Bear", map[string]int{CounterPlusOne: 5})

	g.mu.RLock()
	payable := g.CounterCostPayable(me.ID, src, rc)
	opts := g.CounterCostOptionsForEffect(me.ID, src, rc)
	g.mu.RUnlock()
	if payable {
		t.Fatal("payable with one counter on the source (another permanent's counters don't count)")
	}
	if len(opts) != 1 || opts[0].CardID != src {
		t.Fatalf("options = %+v, want the source alone", opts)
	}

	g.WithWriteLock(func() { _ = g.applyCounterLocked(src, CounterStun, 1) })
	g.mu.RLock()
	payable = g.CounterCostPayable(me.ID, src, rc)
	opts = g.CounterCostOptionsForEffect(me.ID, src, rc)
	g.mu.RUnlock()
	if !payable {
		t.Fatal("not payable with a +1/+1 and a stun counter on the source")
	}
	if len(opts) != 1 || len(opts[0].Kinds) != 2 {
		t.Fatalf("options = %+v, want the source with two kinds", opts)
	}
}
