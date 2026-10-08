package legal

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// cast_mana.go — ADR 0136 §2 and owner answer 1: the total mana a cast
// move charges, on MoveCost.Mana.
//
// A cast's printed `mana_cost` is on the CardView, but it is not what
// the cast pays. CR 601.2f totals the cost from the printed cost or the
// alternative cost the cast claims (CR 118.9), plus the commander tax
// (CR 903.8), the mana half of the additional costs it announces, and
// every increase and reduction on the board. A policy that may not
// import internal/game (ADR 0033 §3) can rebuild none of that. The
// enumerator has already computed it: it is the cost the move's
// payment check was run against. So the move carries it.

// castManaCost renders the mana a cast move charges: `cost` is the cost
// the enumerator's affordability check paid (after the board's
// modifiers, a delve payment and any Phyrexian symbols the move pays
// with life), and `x` the X the move announces.
//
// X is settled into the generic component, because the move announces
// it. A Phyrexian symbol still in the cost is paid with mana by this
// move (the ones it pays with life were struck before the check), so it
// renders as its coloured half. A cast that charges no mana at all (an
// alternative cost of {0}, a free cast permission) renders "{0}" rather
// than "", so a reader can tell "free" from "not stated".
func castManaCost(cost game.ParsedCost, x int) string {
	out := cost
	if out.XSlots > 0 {
		out.Generic += out.XSlots * x
		out.XSlots = 0
	}
	if out.HasPhyrexian || len(out.Required) > 0 {
		req := make([]game.ColorRequirement, len(out.Required))
		for i, r := range out.Required {
			r.Phyrexian = false
			req[i] = r
		}
		out.Required = req
		out.HasPhyrexian = false
	}
	if s := out.String(); s != "" {
		return s
	}
	return "{0}"
}

// withCastMana stamps a cast's mana on a (possibly nil) MoveCost,
// returning a fresh value so no two moves share one.
func withCastMana(c *MoveCost, mana string) *MoveCost {
	out := MoveCost{}
	if c != nil {
		out = *c
		out.Counters = append([]CounterPrice(nil), c.Counters...)
	}
	out.Mana = mana
	return &out
}
