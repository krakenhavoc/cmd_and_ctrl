package legal

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// emerge.go — ADR 0135 §4 (#2416, owner decision 5): which permanent the
// enumerator offers to emerge.
//
// Every other alternative cost costs the same mana whichever cards pay
// its card component, so the enumerator prices the cast once and checks
// each payment only for what it keeps away from the auto-tapper. Emerge
// is the one whose PRICE depends on the payment (CR 702.119a: reduced by
// the sacrificed creature's mana value), so each candidate is priced on
// its own, through the engine's one reduction
// (CostQuery.AltSacrificeManaValue), and only the ones the seat can
// afford are offered: a 1/1 token that saves nothing is not offered
// where only the six-drop makes the cast payable.
//
// The order is the owner's: each candidate's value to the policy (what
// it would rather keep, Options.OrderCostFuel) less the generic mana it
// saves (TargetCandidate.Saves, priced by the policy in its own units),
// cheapest first, and the cap's best are kept. With no policy the
// candidates are ranked by the mana they save alone.

// emergePayments returns the payments an emerge offer is enumerated
// with: one permanent each, affordable at its own price, best first, at
// most maxEnumeratedCostPayments of them. Empty when none is
// affordable, and the cast is then not offered at all.
//
// `base` is the cast's cost before the board's modifiers
// (CastPrice.Base), priced here once per candidate exactly as the
// expansion prices it, so the payment offered first is one the
// expansion's own affordability check then accepts (#544).
func (e *enumerator) emergePayments(
	card game.Card,
	offer *game.AlternativeCost,
	pool []uuid.UUID,
	base game.ParsedCost,
	fromZone game.ZoneKind,
	spend game.ManaSpendContext,
	xFloor int,
) [][]uuid.UUID {
	query := func(mv int) game.CostQuery {
		return game.CostQuery{Card: card, Controller: e.seat, FromZone: fromZone, AltSacrificeManaValue: mv}
	}
	full, err := e.g.ApplyCostModifiersForEffect(base, query(0))
	if err != nil {
		return nil
	}
	type candidate struct {
		id    uuid.UUID
		score float64
	}
	var ranked []candidate
	for _, id := range pool {
		mv := e.g.AltSacrificeManaValueForEffect(id)
		priced, err := e.g.ApplyCostModifiersForEffect(base, query(mv))
		if err != nil {
			continue
		}
		// What the auto-tapper may not spend with THIS permanent named:
		// it may still tap it for mana first (owner decision 3).
		pay := []uuid.UUID{id}
		if !e.canPayExcluding(priced, xFloor, spend, game.CastAutoTapExclusions(game.CastSpellParams{AltCostIDs: pay}, offer)) {
			continue
		}
		// #2701: what the reduction could not take off the printed
		// generic it takes off the X, which buys a larger X for the
		// same mana, so it is saved too.
		saves := full.Generic - priced.Generic + priced.XReduced - full.XReduced
		score := -float64(saves)
		if e.opts.OrderCostFuel != nil {
			score = e.opts.OrderCostFuel(TargetCandidate{ID: id, Saves: saves})
		}
		ranked = append(ranked, candidate{id: id, score: score})
	}
	sort.SliceStable(ranked, func(i, j int) bool { return ranked[i].score < ranked[j].score })
	limit := e.capOr(maxEnumeratedCostPayments)
	if len(ranked) > limit {
		e.noteCut(CapCostPayments, len(ranked)-limit, false)
		ranked = ranked[:limit]
	}
	out := make([][]uuid.UUID, 0, len(ranked))
	for _, c := range ranked {
		out = append(out, []uuid.UUID{c.id})
	}
	return out
}

// emergeManaValue is the reduction a payment buys under `offer`: the one
// named permanent's mana value for an emerge offer, and 0 for every
// other offer and for no payment.
func emergeManaValue(g *game.Game, offer *game.AlternativeCost, paid []uuid.UUID) int {
	if offer == nil || !offer.ReducedBySacrificedManaValue || len(paid) != 1 {
		return 0
	}
	return g.AltSacrificeManaValueForEffect(paid[0])
}

// altPaymentCost is the mana an alternative payment is offered at
// against the first announcement (#1013): the first announcement's own
// cost for every offer whose price does not depend on its payment, and
// for an emerge offer that cost repriced with this payment's reduction
// (ADR 0135 §4). ok is false when the payment cannot be repriced here —
// an announcement that also paid Phyrexian symbols with life or delved,
// which no emerge card prints — so the payment is not offered rather
// than offered at a guess.
func (e *enumerator) altPaymentCost(card game.Card, fromZone game.ZoneKind, first *announcedCast, paid []uuid.UUID, perTarget, varSac bool) (game.ParsedCost, bool) {
	if first.offer == nil || !first.offer.ReducedBySacrificedManaValue {
		return first.cost, true
	}
	if first.life > 0 || len(first.delve) > 0 {
		return game.ParsedCost{}, false
	}
	q := game.CostQuery{
		Card:                  card,
		Controller:            e.seat,
		FromZone:              fromZone,
		XValue:                first.x,
		AltSacrificeManaValue: emergeManaValue(e.g, first.offer, paid),
	}
	if perTarget {
		q.Targets = first.targets
	}
	if varSac {
		q.Sacrificing = len(first.sacs)
	}
	priced, err := e.g.ApplyCostModifiersForEffect(first.base, q)
	if err != nil {
		return game.ParsedCost{}, false
	}
	return priced, true
}
