package game

import (
	"testing"

	"github.com/google/uuid"
)

// payment_facts_test.go — ADR 0109 §8 and §9: a payment fact reaches
// the effect and the targets. The activation-path halves (a discard
// recorded on an ability, a counters-removed bound) are in
// cards/effects/payment_facts_test.go, against registered abilities.

// CR 707.10: a copy of an ability uses the objects used to pay the
// original's costs — every one the record names, each in a list of its
// own.
func TestAbilityCopyCarriesEveryObjectTheCostMoved(t *testing.T) {
	discarded, exiled, attacked := uuid.New(), uuid.New(), uuid.New()
	p := PaidCost{
		Discarded:         []uuid.UUID{discarded},
		Exiled:            []uuid.UUID{exiled},
		Sacrificed:        2,
		ReturnedAttacking: attacked,
		CountersRemoved:   3,
	}
	c := copiedPaidCost(p)
	if len(c.Discarded) != 1 || c.Discarded[0] != discarded {
		t.Errorf("copy's Discarded = %v, want [%v]", c.Discarded, discarded)
	}
	if len(c.Exiled) != 1 || c.Exiled[0] != exiled {
		t.Errorf("copy's Exiled = %v, want [%v]", c.Exiled, exiled)
	}
	if c.Sacrificed != 2 || c.ReturnedAttacking != attacked || c.CountersRemoved != 3 {
		t.Errorf("copy's counts = %+v", c)
	}
	c.Discarded[0], c.Exiled[0] = uuid.Nil, uuid.Nil
	if p.Discarded[0] != discarded || p.Exiled[0] != exiled {
		t.Error("the copy's lists alias the original's")
	}
	if !copiedPaidCost(PaidCost{}).IsZero() {
		t.Error("a copy of an empty record is not empty")
	}
}

// ADR 0109 §9: power and toughness bounds bind like the mana-value
// one — unbound admits everything, bound admits the statistic at most
// X, read with counters.
func TestPowerAndToughnessBoundsBindToTheAnnouncedX(t *testing.T) {
	small := Card{Power: 2, Toughness: 5}
	big := Card{Power: 4, Toughness: 1}
	pumped := Card{Power: 2, Toughness: 2, Counters: map[string]int{"+1/+1": 2}}
	power := (&TargetSpec{Zones: []ZoneKind{ZoneBattlefield}, Min: 1, Max: 1}).WithPowerAtMostX()
	if !power.xBoundAdmits(big) {
		t.Error("before X is bound, the clause admits every card")
	}
	steps := AnnouncedClauses(power, nil, nil)
	bindStepsX(steps, 3)
	if !steps[0].Clause.xBoundAdmits(small) || steps[0].Clause.xBoundAdmits(big) {
		t.Error("bound to X=3: power 2 qualifies, power 4 does not")
	}
	if steps[0].Clause.xBoundAdmits(pumped) {
		t.Error("a 2/2 with two +1/+1 counters has power 4")
	}
	tough := (&TargetSpec{Zones: []ZoneKind{ZoneBattlefield}, Min: 1, Max: 1}).WithToughnessAtMostX()
	steps = AnnouncedClauses(tough, nil, nil)
	bindStepsX(steps, 3)
	if steps[0].Clause.xBoundAdmits(small) || !steps[0].Clause.xBoundAdmits(big) {
		t.Error("bound to X=3: toughness 1 qualifies, toughness 5 does not")
	}
	if power.xBoundSet || tough.xBoundSet {
		t.Error("binding writes the announcement's copy, never the catalog spec")
	}
}

// A clause bounded by the counters removed reads THAT number, not the
// announced X, and an X-bounded clause beside it still reads X.
func TestCountersRemovedBoundReadsTheCounters(t *testing.T) {
	byCounters := (&TargetSpec{Zones: []ZoneKind{ZoneBattlefield}, Min: 1, Max: 1}).WithPowerAtMostX().BoundByTheCountersRemoved()
	byX := (&TargetSpec{Zones: []ZoneKind{ZoneBattlefield}, Min: 1, Max: 1}).WithPowerAtMostX()
	three := Card{Power: 3}
	cs := AnnouncedClauses(byCounters, nil, nil)
	xs := AnnouncedClauses(byX, nil, nil)
	b := AnnouncedBound{X: 1, CountersRemoved: 3}
	bindStepsBound(cs, b)
	bindStepsBound(xs, b)
	if !cs[0].Clause.xBoundAdmits(three) {
		t.Error("three counters removed admit power 3")
	}
	if xs[0].Clause.xBoundAdmits(three) {
		t.Error("an X-bounded clause read the counters instead of X=1")
	}
	if !StepsBoundByCountersRemoved(cs) || StepsBoundByCountersRemoved(xs) {
		t.Error("StepsBoundByCountersRemoved misreports")
	}
}
