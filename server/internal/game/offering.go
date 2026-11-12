package game

import "github.com/google/uuid"

// offering.go — #5657: offering, CR 702.48.
//
// "[Quality] offering" means "As an additional cost to cast this spell,
// you may sacrifice a [quality] permanent. If you chose to pay the
// additional cost, this spell's total cost is reduced by the sacrificed
// permanent's mana cost, and you may cast this spell any time you could
// cast an instant" (CR 702.48a).
//
// Two of the three halves were already costs the engine pays: the
// sacrifice is an optional cost's Sacrifice clause (ADR 0073, the shape
// Gatekeeper of Malakir's kicker has), and the timing is
// AdditionalCost.AsThoughFlash (CR 601.3c, Quantum Reduction's shape),
// which opens the instant-speed window only for a cast that announces
// the cost. What this file adds is the price. CR 702.48b chooses the
// permanent as the spell's choices are made (601.2b) and CR 601.2f
// totals the cost while it is still on the battlefield, so the
// reduction is read off the named permanent here and taken off by
// applyCostModifiersLocked with the other reductions, through
// reduceByManaCost, which is CR 702.48c's and 118.7's rule: generic
// reduces generic, a coloured or colourless symbol reduces mana of the
// same type, and any excess reduces generic.
//
// Emerge (emerge.go) is the neighbour and differs in each half: it is
// an alternative cost, it reduces generic mana by a mana VALUE, and it
// says nothing about timing.

// SacrificedManaCostForEffect is CostQuery.SacrificedManaCost for one
// announcement: the mana cost of each permanent `sacrificeIDs` names to
// a cost that is ReducedBySacrificedManaCost, summed. `mandatory` and
// `optional` are the card's costs and `chosen` the announced optional
// indices; the flat sacrifice list is walked in the plan's order, as
// validateAdditionalCostLocked walks it, so the permanent read is the
// one the validator holds to that cost's clause.
//
// The zero cost when no announced cost has the flag, and when the list
// is too short for the plan (the validator refuses that cast; this only
// keeps a malformed announcement from being priced as a discount).
//
// Exported for the bot enumerator, which prices each payment before
// announcing it, so it reads the number the payment is charged with
// (#544). Caller must hold g.mu.
func (g *Game) SacrificedManaCostForEffect(mandatory *AdditionalCost, optional []AdditionalCost, chosen []int, sacrificeIDs []uuid.UUID) ParsedCost {
	var by ParsedCost
	if len(sacrificeIDs) == 0 {
		return by
	}
	si := 0
	for _, pay := range castCostPayments(mandatory, optional, chosen) {
		if pay.cost.Sacrifice == nil {
			continue
		}
		n := len(sacrificeIDs) - si
		if !SacrificeCostVariable(pay.cost.Sacrifice) {
			n = SacrificeCostCount(pay.cost.Sacrifice)
		}
		if n < 0 || si+n > len(sacrificeIDs) {
			return ParsedCost{}
		}
		if pay.cost.ReducedBySacrificedManaCost {
			for _, id := range sacrificeIDs[si : si+n] {
				mc := g.permanentManaCostLocked(id)
				by.Generic += mc.Generic
				by.Required = append(by.Required, mc.Required...)
			}
		}
		si += n
	}
	return by
}

// AnySacrificeReducesByManaCost reports whether the announcement pays a
// cost that is ReducedBySacrificedManaCost: what tells the enumerator
// that this cast's price depends on which permanent pays it.
func AnySacrificeReducesByManaCost(optional []AdditionalCost, chosen []int) bool {
	for _, i := range chosen {
		if i >= 0 && i < len(optional) && optional[i].ReducedBySacrificedManaCost {
			return true
		}
	}
	return false
}

// permanentManaCostLocked is the mana cost of the permanent `id` as it
// stands on the battlefield (CR 202.1): none for a token that isn't a
// copy and for a transformed back face (CR 202.1b: Card.ManaCost is
// empty), none for a face-down permanent (CR 708.2), and an {X} in it
// is 0 (CR 107.3g), so it reduces nothing. A mana cost the engine can't
// read reduces nothing, which errs toward the printed price. The zero
// cost when `id` names no permanent. Caller must hold g.mu.
func (g *Game) permanentManaCostLocked(id uuid.UUID) ParsedCost {
	if g.Battlefield == nil {
		return ParsedCost{}
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.InstanceID != id {
			continue
		}
		if c.FaceDown || c.ManaCost == "" {
			return ParsedCost{}
		}
		mc, err := ParseCost(c.ManaCost)
		if err != nil {
			return ParsedCost{}
		}
		return ParsedCost{Generic: mc.Generic, Required: mc.Required}
	}
	return ParsedCost{}
}

// sacrificedManaCostLocked is SacrificedManaCostForEffect for a cast's
// own announcement, with the costs resolved as printedCostLocked
// resolves them. Caller must hold g.mu.
func (g *Game) sacrificedManaCostLocked(playerID uuid.UUID, card Card, params CastSpellParams) ParsedCost {
	if len(params.SacrificeIDs) == 0 || len(params.OptionalCosts) == 0 {
		return ParsedCost{}
	}
	optional := g.OptionalCostsForCastLocked(playerID, card)
	if !AnySacrificeReducesByManaCost(optional, params.OptionalCosts) {
		return ParsedCost{}
	}
	mandatory, err := ChosenAdditionalCost(AdditionalCostFor(CatalogKey(card)), params.CostBranch)
	if err != nil {
		mandatory = nil
	}
	return g.SacrificedManaCostForEffect(mandatory, optional, params.OptionalCosts, params.SacrificeIDs)
}
