package game

// counter_cost_x.go — one announced X for an {X} mana symbol and a
// "Remove X counters" cost in the same activation cost (#6085).
//
//	Talon of Pain  {X}, {T}, Remove X charge counters from this
//	               artifact: It deals X damage to any target.
//
// CR 107.3a: "While an activated ability is on the stack, any X in its
// activation cost equals the announced value." The activator announces
// X once (CR 601.2b by way of 602.2b), and that one number is both the
// generic mana the {X} charges and the counters the removal takes. A
// plain RemoveCountersXFromThis announces its count on its own
// (ActivateAbilityParams.CounterCounts), which is right for a cost with
// no other X (Geistflame Reservoir's "any number") and wrong here: two
// numbers would let a Talon pay {1} and remove five counters.
//
// CounterRemovalCost.CountIsX ties them. The count IS
// ActivateAbilityParams.XValue: a payment that sends no count is read
// as X, one that sends a count must send exactly X, and anything else
// is refused before a cost is paid. X = 0 is the empty payment
// (announcesNoCounters), which is legal and removes nothing. An X
// above the counters the source holds fails the ordinary per-part
// check (ErrInsufficientCounters), so the counters bound X from above
// exactly as the pool does.
//
// The protocol view projects no counter picker for the component (the
// X picker is the payment), and the legal-move enumerator offers the
// largest X both the pool and the counters can pay
// (internal/legal/counter_cost_x.go).

import "github.com/google/uuid"

// CounterCountIsX reports a removal whose count is the activation's
// announced X (CounterRemovalCost.CountIsX). Nil-safe.
func CounterCountIsX(rc *CounterRemovalCost) bool {
	return rc != nil && rc.CountIsX
}

// counterPaymentAtX is the announce-time payment for a CountIsX removal
// with its count fixed to the announced X: the count the client sent,
// when it agrees, or X when it sent none. A count that disagrees with X,
// or a split into more than one part, is ErrInvalidParam — one announced
// X cannot be two numbers (CR 107.3a). Any other cost is returned
// unchanged.
func counterPaymentAtX(rc *CounterRemovalCost, sourceID uuid.UUID, pay CounterCostPayment, x int) (CounterCostPayment, error) {
	if !CounterCountIsX(rc) {
		return pay, nil
	}
	if len(pay.Counts) > 1 || (len(pay.Counts) == 1 && pay.Counts[0] != x) {
		return pay, ErrInvalidParam
	}
	if len(pay.SourceIDs) > 1 || (len(pay.SourceIDs) == 1 && pay.SourceIDs[0] != sourceID) {
		return pay, ErrInvalidParam
	}
	pay.Counts = []int{x}
	return pay, nil
}
