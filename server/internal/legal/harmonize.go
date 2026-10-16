package legal

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// harmonize.go — #3931 (CR 702.180, ADR 0135 amendment 2026-10-16):
// which creature the enumerator offers to tap for harmonize.
//
// Emerge's discipline (emerge.go) with a tap for a sacrifice: the price
// depends on the payment (reduced by the tapped creature's power), so
// each candidate is priced on its own through the engine's one
// reduction (CostQuery.AltTappedPower), and only the affordable ones are
// offered, ranked by what tapping the creature costs the policy less
// the generic mana it saves (TargetCandidate.Tap and Saves). Unlike
// emerge, naming nothing is a legal payment: it is offered last, when
// the seat can afford the full harmonize cost. A creature whose power
// saves nothing is never offered, since tapping nothing pays the same.

// harmonizePayments returns the payments a harmonize offer is
// enumerated with: one creature each, affordable at its own price and
// saving mana, best first, then the empty payment when it is affordable;
// at most maxEnumeratedCostPayments in all. Empty when none is
// affordable, and the cast is then not offered at all.
//
// `base` is the cast's cost before the board's modifiers, priced here
// once per candidate exactly as the expansion prices it (#544).
func (e *enumerator) harmonizePayments(
	card game.Card,
	offer *game.AlternativeCost,
	pool []uuid.UUID,
	base game.ParsedCost,
	fromZone game.ZoneKind,
	spend game.ManaSpendContext,
	xFloor int,
) [][]uuid.UUID {
	query := func(power int) game.CostQuery {
		return game.CostQuery{Card: card, Controller: e.seat, FromZone: fromZone, AltTappedPower: power}
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
		priced, err := e.g.ApplyCostModifiersForEffect(base, query(e.g.AltTappedPowerForEffect(id)))
		if err != nil {
			continue
		}
		// #2701: what the reduction could not take off the printed
		// generic it takes off the X, which buys a larger X.
		saves := full.Generic - priced.Generic + priced.XReduced - full.XReduced
		if saves < 1 {
			continue
		}
		// The tapped creature cannot also make the mana (CR 601.2g
		// before 601.2h: it is tapped for the cost, not for mana).
		pay := []uuid.UUID{id}
		if !e.canPayExcluding(priced, xFloor, spend, game.CastAutoTapExclusions(game.CastSpellParams{AltCostIDs: pay}, offer)) {
			continue
		}
		score := -float64(saves)
		if e.opts.OrderCostFuel != nil {
			score = e.opts.OrderCostFuel(TargetCandidate{ID: id, Tap: true, Saves: saves})
		}
		ranked = append(ranked, candidate{id: id, score: score})
	}
	sort.SliceStable(ranked, func(i, j int) bool { return ranked[i].score < ranked[j].score })
	limit := e.capOr(maxEnumeratedCostPayments)
	none := e.canPayExcluding(full, xFloor, spend, nil)
	keep := limit
	if none {
		keep--
	}
	if len(ranked) > keep {
		e.noteCut(CapCostPayments, len(ranked)-keep, false)
		ranked = ranked[:max(keep, 0)]
	}
	out := make([][]uuid.UUID, 0, len(ranked)+1)
	for _, c := range ranked {
		out = append(out, []uuid.UUID{c.id})
	}
	if none {
		out = append(out, nil)
	}
	return out
}

// harmonizePower is the reduction a payment buys under `offer`: the one
// named creature's power for a harmonize offer, and 0 for every other
// offer and for no payment.
func harmonizePower(g *game.Game, offer *game.AlternativeCost, paid []uuid.UUID) int {
	if offer == nil || !offer.ReducedByTappedPower || len(paid) != 1 {
		return 0
	}
	return g.AltTappedPowerForEffect(paid[0])
}
