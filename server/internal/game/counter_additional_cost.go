package game

import "github.com/google/uuid"

// counter_additional_cost.go — #6121: "As an additional cost to cast this
// spell, you may remove any number of +1/+1 counters from among creatures
// you control. This spell costs {2} less to cast for each counter removed
// this way." (Hierophant Bio-Titan's Frenzied Metabolism).
//
// The component is AdditionalCost.RemoveCounters: the activated ability's
// counter cost (#625, #789) on a cast's mandatory slot, validated and paid
// through the same validateCounterRemovalLocked / payCounterRemovalLocked
// an alternative cost's counters go through (#5390). CR 601.2b announces
// the payment, CR 601.2f totals the cost with the announced count (the
// spell's own cost modifier reads CostQuery.RemovingCounters), and CR
// 601.2h removes the counters with the spell on the stack.

// planRemoveCounters is the counter component of the announced plan's
// mandatory cost, or nil. Register keeps the component on the mandatory
// slot, so an optional cost never carries one.
func planRemoveCounters(plan []costPayment) *CounterRemovalCost {
	for _, pay := range plan {
		if pay.index == -1 && pay.cost.RemoveCounters != nil {
			return pay.cost.RemoveCounters
		}
	}
	return nil
}

// additionalCostRemovesCounters reports whether the card's catalog
// additional cost has a counter component. The alternative cost's
// validator reads the shared counter lists only when it does not, so
// neither component claims the other's payment.
func additionalCostRemovesCounters(card Card) bool {
	ac := AdditionalCostFor(CatalogKey(card))
	return ac != nil && ac.RemoveCounters != nil
}

// validateAdditionalCostCountersLocked resolves the mandatory cost's
// counter payment, removing nothing (validate-all-then-pay, ADR 0020 §3).
//
// The payment rides the lists an alternative cost's counters ride. A cast
// that claims an offer with its own counter component spends them there,
// and this component is paid with none, which its zero floor allows. A
// payment that names nothing is the announcement of zero: "you may remove
// any number" includes none (CR 601.2b). Anything else goes through the
// activation's validator with the spell being cast as the source (the
// cast is not on the battlefield, so no creature is it).
//
// Caller must hold g.mu.
func (g *Game) validateAdditionalCostCountersLocked(playerID, castID uuid.UUID, plan []costPayment, alt *AlternativeCost, pay CounterCostPayment) (counterPayment, error) {
	rc := planRemoveCounters(plan)
	if rc == nil || (alt != nil && alt.RemoveCounters != nil) {
		return counterPayment{}, nil
	}
	if pay.empty() && rc.Variable && rc.N == 0 {
		return counterPayment{}, nil
	}
	return g.validateCounterRemovalLocked(playerID, castID, rc, pay)
}

// castAltCounterPayment is the counter payment the alternative cost's
// validator reads: the announcement's lists, unless the card's own
// additional cost is the counter component they pay (#6121) and the
// claimed offer has none.
func castAltCounterPayment(card Card, alt *AlternativeCost, params CastSpellParams) CounterCostPayment {
	if (alt == nil || alt.RemoveCounters == nil) && additionalCostRemovesCounters(card) {
		return CounterCostPayment{}
	}
	return castCounterPayment(params)
}

// castCountersAnnouncedLocked is CostQuery.RemovingCounters for a cast's
// own announcement: the counters named to the card's additional counter
// cost, as the counts sum. Zero when the card has none. effects.Register
// refuses a card whose own alternative cost removes counters beside it,
// so the lists are this component's. The validator holds the counts to
// the board; this only prices them (CR 601.2f).
//
// Caller must hold g.mu.
func (g *Game) castCountersAnnouncedLocked(card Card, params CastSpellParams) int {
	if !additionalCostRemovesCounters(card) {
		return 0
	}
	return countsTotal(params.CounterCounts)
}

// CastCountersAnnouncedForAdditionalCost is castCountersAnnouncedLocked's
// arithmetic for a caller building an announcement (the bot enumerator):
// the counts the payment names, summed. Negative counts read as zero; the
// validator refuses them.
func CastCountersAnnouncedForAdditionalCost(counts []int) int { return countsTotal(counts) }

// countsTotal sums a counter payment's counts, a negative one as zero.
func countsTotal(counts []int) int {
	n := 0
	for _, c := range counts {
		if c > 0 {
			n += c
		}
	}
	return n
}
