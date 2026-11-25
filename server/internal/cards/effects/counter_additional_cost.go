package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// counter_additional_cost.go — #6121: removing counters as a spell's
// additional cost, with a discount per counter removed.
//
//	Hierophant Bio-Titan   "Frenzied Metabolism — As an additional cost to
//	                       cast this spell, you may remove any number of
//	                       +1/+1 counters from among creatures you control.
//	                       This spell costs {2} less to cast for each
//	                       counter removed this way."
//
// is
//
//	AdditionalCost: RemoveAnyNumberOfCountersAmongCost("+1/+1", "creatures you control", Creature()),
//	SelfCostModifiers: []game.CostModifier{CostsLessPerCounterRemoved("{2}", "Frenzied Metabolism")},
//
// The component is game.AdditionalCost.RemoveCounters; the payment rides
// cast_spell's counter_source_ids / counter_counts, CR 601.2b announces
// it, CR 601.2f prices the cast with the announced total
// (game.CostQuery.RemovingCounters) and CR 601.2h removes the counters
// with the spell on the stack. "You may … any number" includes none, so
// the floor is zero and a cast that names no counters pays full price.

// RemoveAnyNumberOfCountersAmongCost is "remove any number of <kind>
// counters from among <permanents> you control" on a cast's mandatory
// slot. The permanents are chosen, not targeted, and "you control" is the
// engine's rule.
func RemoveAnyNumberOfCountersAmongCost(kind, label string, preds ...CardPredicate) *game.AdditionalCost {
	return &game.AdditionalCost{
		Label: "Remove any number of " + kind + " counters from among " + label,
		RemoveCounters: &game.CounterRemovalCost{
			Counter:  kind,
			From:     TargetPermanent(label, preds...),
			Among:    true,
			Variable: true,
		},
	}
}

// CostsLessPerCounterRemoved is "This spell costs {per} less to cast for
// each counter removed this way" (#6121): it reads
// game.CostQuery.RemovingCounters, the total the announcement names to the
// card's RemoveAnyNumberOfCountersAmongCost. A reduction spends generic
// mana only and stops at zero (CR 601.2f), so Hierophant Bio-Titan always
// costs at least its {G}{G}. `per` must be generic mana; anything else
// panics at init.
func CostsLessPerCounterRemoved(per string, label string) game.CostModifier {
	unit, err := game.ParseCost(per)
	if err != nil || unit.XSlots > 0 || len(unit.Required) > 0 || unit.Generic < 1 {
		panic(fmt.Sprintf("effects.CostsLessPerCounterRemoved: %q is not a fixed generic amount (%v)", per, err))
	}
	return CostsLessEach(func(q game.CostQuery) int {
		if q.RemovingCounters < 1 {
			return 0
		}
		return unit.Generic * q.RemovingCounters
	}, label)
}

// checkAdditionalCounterCost is Register's guard for a cast's counter
// component (#6121). It lives on the mandatory slot only, in the one
// shape the validator and the view read: a printed kind, removed from
// among permanents a "from" clause matches, in any number from zero. An
// optional cost and an either/or branch can't carry it, and neither can a
// card whose own alternative cost removes counters, since both would ride
// the same counter lists.
func checkAdditionalCounterCost(spec Spec) {
	for i, oc := range spec.OptionalCosts {
		if oc.RemoveCounters != nil {
			panic(fmt.Sprintf("effects.Register: %q optional cost %d removes counters — the mandatory slot's alone (#6121)", spec.Name, i))
		}
	}
	ac := spec.AdditionalCost
	if ac == nil {
		return
	}
	for _, b := range ac.Either {
		if b.RemoveCounters != nil {
			panic(fmt.Sprintf("effects.Register: %q has an either/or branch that removes counters — the mandatory slot's alone (#6121)", spec.Name))
		}
	}
	rc := ac.RemoveCounters
	if rc == nil {
		return
	}
	if rc.Counter == "" || rc.From == nil || !rc.Among || !rc.Variable || rc.N != 0 || rc.All || rc.OnePermanent || len(rc.Kinds) > 0 {
		panic(fmt.Sprintf("effects.Register: %q removes counters as an additional cost in a shape no reader takes — build it with RemoveAnyNumberOfCountersAmongCost (#6121)", spec.Name))
	}
	for _, alt := range spec.AlternativeCosts {
		if alt.RemoveCounters != nil {
			panic(fmt.Sprintf("effects.Register: %q removes counters both as an additional cost and in alternative cost %q — they share the counter lists (#6121)", spec.Name, alt.Key))
		}
	}
}
