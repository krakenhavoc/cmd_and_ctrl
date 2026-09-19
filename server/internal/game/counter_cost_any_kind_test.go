package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// counter_cost_any_kind_test.go — #943: the last printed shape of the
// counter-cost component, an ANY-KIND removal split across permanents.
//
// Tekuthal, Inquiry Dominus' "Remove three counters from among other
// artifacts, creatures, and planeswalkers you control" takes three
// counters of whatever kinds happen to be there, so the payment names
// a kind PER PERMANENT as well as a count. Everything else about it is
// #789's among cost: one component, one validator, one candidate walk,
// validated as a set, paid at announce, never replaceable.
//
// What is pinned here is exactly the difference: the kinds may vary
// across the parts, one permanent may pay in two of them, and every
// refusal the fixed-kind form has still refuses.

// otherPermanentsCostSpec is Tekuthal's clause as a cost predicate:
// artifacts, creatures and planeswalkers you control other than the
// named source. "Other" by name, the posture every "another" cost
// clause in the catalog takes — a cost clause is built before any
// instance exists.
func otherPermanentsCostSpec(exclude string) *TargetSpec {
	return &TargetSpec{
		Mode:  "permanent",
		Label: "other artifacts, creatures, and planeswalkers you control",
		Zones: []ZoneKind{ZoneBattlefield},
		CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
			return c.Name != exclude && (c.IsArtifact() || c.IsCreature() || c.IsPlaneswalker())
		},
		Min: 1, Max: 1,
	}
}

// tekuthalCost is the shape under test: three counters of any kind,
// from among other permanents you control.
func tekuthalCost() AbilityCost {
	return AbilityCost{RemoveCounters: &CounterRemovalCost{
		N:     3,
		From:  otherPermanentsCostSpec("Counter Cost Source"),
		Among: true,
	}}
}

// The headline: a loyalty counter off a planeswalker and two +1/+1
// counters off a creature is ONE legal payment of "remove three
// counters from among …".
func TestAnyKindAmongCounterCostMixesKinds(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushCounterCostSource(g, me, tekuthalCost(), nil)
	bear := pushPermanentWithCounters(g, me, "Bear", "Creature — Bear", map[string]int{"+1/+1": 2})
	jace := pushPermanentWithCounters(g, me, "Jace", "Planeswalker — Jace", map[string]int{CounterLoyalty: 4})

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{
		CounterSourceIDs: []uuid.UUID{bear, jace},
		CounterCounts:    []int{2, 1},
		CounterKinds:     []string{CounterPlusOne, CounterLoyalty},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := counterOf(g, bear, CounterPlusOne); got != 0 {
		t.Errorf("Bear has %d +1/+1 counters, want 0", got)
	}
	if got := counterOf(g, jace, CounterLoyalty); got != 3 {
		t.Errorf("Jace has %d loyalty counters, want 3", got)
	}
	// CR 606.3 is about a walker's OWN abilities: paying someone
	// else's cost with a loyalty counter is not a loyalty activation.
	if len(g.LoyaltyActivatedThisTurn) != 0 {
		t.Errorf("paying a cost with a loyalty counter burned the walker's turn: %v", g.LoyaltyActivatedThisTurn)
	}
}

// One permanent may pay in TWO kinds — which is why the payment's set
// is keyed on (permanent, kind) rather than on the permanent.
func TestAnyKindAmongCounterCostTakesTwoKindsFromOnePermanent(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushCounterCostSource(g, me, tekuthalCost(), nil)
	bear := pushPermanentWithCounters(g, me, "Bear", "Creature — Bear",
		map[string]int{"+1/+1": 2, CounterShield: 1})

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{
		CounterSourceIDs: []uuid.UUID{bear, bear},
		CounterCounts:    []int{2, 1},
		CounterKinds:     []string{CounterPlusOne, CounterShield},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := counterOf(g, bear, CounterPlusOne); got != 0 {
		t.Errorf("Bear has %d +1/+1 counters, want 0", got)
	}
	if got := counterOf(g, bear, CounterShield); got != 0 {
		t.Errorf("Bear has %d shield counters, want 0", got)
	}
}

// Every refusal, and the rule that a refused payment removes nothing
// (ADR 0020 §3: validate everything, then pay everything).
func TestAnyKindAmongCounterCostIsValidatedAsASet(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushCounterCostSource(g, me, tekuthalCost(), nil)
	bear := pushPermanentWithCounters(g, me, "Bear", "Creature — Bear", map[string]int{"+1/+1": 2})
	jace := pushPermanentWithCounters(g, me, "Jace", "Planeswalker — Jace", map[string]int{CounterLoyalty: 4})
	land := pushPermanentWithCounters(g, me, "Vivid Creek", "Land", map[string]int{"charge": 3})
	theirs := pushPermanentWithCounters(g, g.Seats[1], "Their Bear", "Creature — Bear", map[string]int{"+1/+1": 5})

	cases := []struct {
		name   string
		ids    []uuid.UUID
		counts []int
		kinds  []string
		kind   string
		want   error
	}{
		{"a mixed payment over the printed total",
			[]uuid.UUID{bear, jace}, []int{2, 2}, []string{CounterPlusOne, CounterLoyalty}, "", ErrInvalidParam},
		{"a mixed payment short of the printed total",
			[]uuid.UUID{bear, jace}, []int{1, 1}, []string{CounterPlusOne, CounterLoyalty}, "", ErrInvalidParam},
		{"a kind the permanent does not hold",
			[]uuid.UUID{bear, jace}, []int{2, 1}, []string{CounterPlusOne, CounterShield}, "", ErrInsufficientCounters},
		{"more of a kind than the permanent holds",
			[]uuid.UUID{bear}, []int{3}, []string{CounterPlusOne}, "", ErrInsufficientCounters},
		{"a part that removes nothing",
			[]uuid.UUID{jace, bear}, []int{3, 0}, []string{CounterLoyalty, CounterPlusOne}, "", ErrInvalidParam},
		{"the same permanent and kind twice",
			[]uuid.UUID{bear, bear, jace}, []int{1, 1, 1}, []string{CounterPlusOne, CounterPlusOne, CounterLoyalty}, "", ErrInvalidParam},
		{"the source, which the clause excludes",
			[]uuid.UUID{src, jace}, []int{2, 1}, []string{CounterPlusOne, CounterLoyalty}, "", ErrIllegalTarget},
		{"a permanent the clause does not match",
			[]uuid.UUID{land, bear}, []int{1, 2}, []string{"charge", CounterPlusOne}, "", ErrIllegalTarget},
		{"a permanent another player controls",
			[]uuid.UUID{theirs}, []int{3}, []string{CounterPlusOne}, "", ErrCardCallerMismatch},
		{"kinds that do not line up with the ids",
			[]uuid.UUID{bear, jace}, []int{2, 1}, []string{CounterPlusOne}, "", ErrInvalidParam},
		{"an empty kind in the array",
			[]uuid.UUID{bear, jace}, []int{2, 1}, []string{CounterPlusOne, ""}, "", ErrInvalidParam},
		{"no kind named at all",
			[]uuid.UUID{bear, jace}, []int{2, 1}, nil, "", ErrInvalidParam},
		{"kinds that contradict the single kind sent beside them",
			[]uuid.UUID{bear, jace}, []int{2, 1}, []string{CounterPlusOne, CounterLoyalty}, CounterPlusOne, ErrInvalidParam},
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
			if counterOf(g, bear, CounterPlusOne) != 2 || counterOf(g, jace, CounterLoyalty) != 4 ||
				counterOf(g, land, "charge") != 3 {
				t.Fatal("a refused payment removed counters")
			}
		})
	}
}

// A payment whose parts happen to share a kind still speaks the #625
// wire shape — one counter_kind, no array — and the engine reads it
// the same way. This is what keeps every client that predates #943
// working against an any-kind among cost.
func TestAnyKindAmongCounterCostAcceptsASingleKind(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushCounterCostSource(g, me, tekuthalCost(), nil)
	a := pushPermanentWithCounters(g, me, "Bear", "Creature — Bear", map[string]int{"+1/+1": 2})
	b := pushPermanentWithCounters(g, me, "Servo", "Artifact Creature — Servo", map[string]int{"+1/+1": 1})

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{
		CounterSourceIDs: []uuid.UUID{a, b},
		CounterCounts:    []int{2, 1},
		CounterKind:      CounterPlusOne,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if counterOf(g, a, CounterPlusOne)+counterOf(g, b, CounterPlusOne) != 0 {
		t.Error("the single-kind payment did not come off")
	}
}

// CounterCostPayable is what greys the menu row and what stops the
// enumerator offering the ability. For an any-kind among cost it has
// to count EVERY kind, because every kind can pay.
func TestAnyKindAmongCounterCostPayableCountsEveryKind(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	rc := tekuthalCost().RemoveCounters
	src := pushCounterCostSource(g, me, tekuthalCost(), nil)

	g.mu.RLock()
	payable := g.CounterCostPayable(me.ID, src, rc)
	g.mu.RUnlock()
	if payable {
		t.Fatal("payable with nothing on the board")
	}

	bear := pushPermanentWithCounters(g, me, "Bear", "Creature — Bear", map[string]int{"+1/+1": 2})
	g.mu.RLock()
	payable = g.CounterCostPayable(me.ID, src, rc)
	g.mu.RUnlock()
	if payable {
		t.Fatal("payable with two counters against a cost of three")
	}

	// A third counter of a DIFFERENT kind completes the payment: the
	// kinds are added together, which a fixed-kind among cost would
	// not do.
	g.WithWriteLock(func() { _ = g.applyCounterLocked(bear, CounterShield, 1) })
	g.mu.RLock()
	payable = g.CounterCostPayable(me.ID, src, rc)
	g.mu.RUnlock()
	if !payable {
		t.Fatal("not payable with two +1/+1 and one shield counter against a cost of three")
	}
}

// Undo across the activation: the rewind puts every kind back where it
// came from, and the same mixed payment can be announced again.
func TestUndoAcrossAnAnyKindAmongCounterPayment(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	// RestoreFrom swaps g.Seats wholesale, so the seat is re-read
	// after the rewind rather than captured once.
	seat := func() *Player { return g.Seats[0] }
	src := pushCounterCostSource(g, seat(), tekuthalCost(), nil)
	bear := pushPermanentWithCounters(g, seat(), "Bear", "Creature — Bear", map[string]int{"+1/+1": 2})
	jace := pushPermanentWithCounters(g, seat(), "Jace", "Planeswalker — Jace", map[string]int{CounterLoyalty: 4})
	pay := ActivateAbilityParams{
		CounterSourceIDs: []uuid.UUID{bear, jace},
		CounterCounts:    []int{2, 1},
		CounterKinds:     []string{CounterPlusOne, CounterLoyalty},
	}

	before := g.Clone()
	if err := g.ActivateCatalogAbility(seat().ID, src, 0, pay); err != nil {
		t.Fatalf("activate: %v", err)
	}
	g.WithWriteLock(func() { g.RestoreFrom(before) })

	if got := counterOf(g, bear, CounterPlusOne); got != 2 {
		t.Errorf("after the rewind Bear has %d +1/+1 counters, want 2", got)
	}
	if got := counterOf(g, jace, CounterLoyalty); got != 4 {
		t.Errorf("after the rewind Jace has %d loyalty counters, want 4", got)
	}
	if g.Stack.Size() != 0 {
		t.Errorf("the ability survived the rewind on the stack")
	}
	if err := g.ActivateCatalogAbility(seat().ID, src, 0, pay); err != nil {
		t.Fatalf("replayed activation: %v", err)
	}
	if got := counterOf(g, jace, CounterLoyalty); got != 3 {
		t.Errorf("after the replay Jace has %d loyalty counters, want 3", got)
	}
}
