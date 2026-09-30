package legal

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// delve.go — ADR 0100 §6 and owner decision 3: which graveyard cards a
// delve cast (CR 702.66a) is offered to exile.
//
// TWO payments, not every subset. The expansion is already modes ×
// targets × cost payments, and the number of cards exiled is a cost
// variable, which ADR 0033 §1's corollary says must not become an arity
// of the target cross product. So a delve cast is offered:
//
//   - the FEWEST graveyard cards that make it affordable, at the
//     largest X the seat can reach — which is the ordinary move, emitted
//     once per target set like any other cast, and the empty payment
//     when the pool already pays; and
//   - the FULL budget, min(DelveBudget, graveyard size) cards, offered
//     once against the first announcement out of the leftover budget,
//     the way #1013 offers an alternative-cost payment. Murktide Regent
//     and Soulflayer want more of the graveyard exiled, not less.
//
// Both take the pool in the seat's fuel order (cheapestFuelFirst), so
// the cards a payment eats first are the ones the policy misses least.
// The arithmetic is the engine's own — game.DelveBudgetFor and
// game.DelveAdjusted over the cost the target set was priced at — so a
// payment offered here is one CastSpell's validator accepts (#544).

// delvePayment is one delve payment: the X it is announced with, the
// cards it exiles, and the cost left for mana once they are exiled.
type delvePayment struct {
	x    int
	ids  []uuid.UUID
	cost game.ParsedCost
}

// castSolve is castPaymentSolve with the delve half: the fewest-cards
// payment the move carries, and the full-budget one (nil when the two
// are the same set, or when the card has no delve).
type castSolve struct {
	castPaymentSolve
	delve     []uuid.UUID
	delveFull *delvePayment
}

// solveCast is castPayment for a cast that may also delve. With no
// delve pool it is castPayment verbatim. With one it solves X and the
// fewest cards together, falling back to castPayment — and its
// Phyrexian life path — when even the full budget cannot make the cast
// affordable, which is no worse than a card with no delve at all.
func (e *enumerator) solveCast(
	priced game.ParsedCost,
	spend game.ManaSpendContext,
	floor, lifeCeiling, reserved int,
	lifeAllowed bool,
	pool []uuid.UUID,
) (castSolve, bool) {
	if len(pool) > 0 {
		if s, ok := e.delvePlan(priced, spend, floor, lifeCeiling, pool); ok {
			return s, true
		}
	}
	pay, ok := e.castPayment(priced, spend, floor, lifeCeiling, reserved, lifeAllowed)
	return castSolve{castPaymentSolve: pay}, ok
}

// delvePlan finds the largest X the seat can pay with the full delve
// budget exiled, then the fewest cards that still pay for that X.
func (e *enumerator) delvePlan(
	priced game.ParsedCost,
	spend game.ManaSpendContext,
	floor, lifeCeiling int,
	pool []uuid.UUID,
) (castSolve, bool) {
	fullAt := func(x int) (int, game.ParsedCost) {
		n := game.DelveBudgetFor(priced, x)
		if n > len(pool) {
			n = len(pool)
		}
		return n, game.DelveAdjusted(priced, n, x)
	}
	lo, hi := 0, 0
	if priced.XSlots > 0 {
		if floor < 0 {
			floor = 0
		}
		lo, hi = floor, e.opts.MaxX
		if lifeCeiling != noXCeiling && lifeCeiling < hi {
			hi = lifeCeiling
		}
	}
	best, fullN := -1, 0
	var fullCost game.ParsedCost
	for x := lo; x <= hi; x++ {
		n, c := fullAt(x)
		if !e.canPay(c, x, spend) {
			break
		}
		best, fullN, fullCost = x, n, c
	}
	if best < 0 {
		return castSolve{}, false
	}
	fewest, fewestCost := fullN, fullCost
	for n := 0; n < fullN; n++ {
		c := game.DelveAdjusted(priced, n, best)
		if e.canPay(c, best, spend) {
			fewest, fewestCost = n, c
			break
		}
	}
	out := castSolve{
		castPaymentSolve: castPaymentSolve{x: best, cost: fewestCost},
		delve:            append([]uuid.UUID(nil), pool[:fewest]...),
	}
	if fullN != fewest {
		out.delveFull = &delvePayment{x: best, ids: append([]uuid.UUID(nil), pool[:fullN]...), cost: fullCost}
	}
	return out, true
}

// delveFullAffordable re-checks the full-budget payment against the
// first announcement's other named payments, which the auto-tapper may
// not spend (game.CastAutoTapExclusions). Exiling more only lowers the
// mana owed, so this is a belt, but the enumerator offers nothing it
// has not priced.
func (e *enumerator) delveFullAffordable(first *announcedCast, spend game.ManaSpendContext) bool {
	full := first.delveFull
	if full == nil {
		return false
	}
	return e.canPayExcluding(full.cost, full.x, spend, game.CastAutoTapExclusions(game.CastSpellParams{
		DiscardIDs:   first.discards,
		SacrificeIDs: first.sacs,
		TeamworkIDs:  first.team,
		BlightIDs:    first.blight,
	}))
}
