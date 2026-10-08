package game

import "github.com/google/uuid"

// emerge.go — ADR 0135 §4 (#2416): emerge, CR 702.119.
//
// "Emerge [cost]" means "You may cast this spell by paying [cost] and
// sacrificing a creature rather than paying its mana cost" and "If you
// chose to pay this spell's emerge cost, its total cost is reduced by an
// amount of generic mana equal to the sacrificed creature's mana value"
// (CR 702.119a). It is an ordinary sacrifice offer
// (AlternativeCost.Sacrifice, #1727) with one more flag,
// ReducedBySacrificedManaValue, so the validator, the payer and the
// picker are the ones Dread Return's flashback already uses. What this
// file adds is the price: CR 702.119c chooses the permanent as the cost
// is chosen (601.2b), and CR 601.2f totals the cost while it is still on
// the battlefield, so the reduction is read off the named permanent here
// and taken off by applyCostModifiersLocked with the other reductions.

// claimedAltCostLocked resolves the alternative cost `params` claims for
// `card`, through the precedence printedCostLocked prices with (the
// card's own offers, then a permission's, then a granted one). Nil when
// the cast claims none, or claims one the card does not offer. Caller
// must hold g.mu.
func (g *Game) claimedAltCostLocked(playerID uuid.UUID, card Card, params CastSpellParams) *AlternativeCost {
	if params.AlternativeCost == "" {
		return nil
	}
	srcKind, ok := castZoneFromWire(params.FromZone)
	if !ok {
		return nil
	}
	grant := g.CastPermissionForClaimLocked(playerID, card, srcKind, params.AlternativeCost)
	alt, err := g.resolveAlternativeCostLocked(playerID, card, srcKind, grant, params.AlternativeCost, nil)
	if err != nil {
		return nil
	}
	return alt
}

// altSacrificeManaValueLocked is CostQuery.AltSacrificeManaValue for one
// announcement: the mana value of the one permanent the claimed emerge
// offer names in AltCostIDs, or 0 when the cast claims no such offer or
// names anything but exactly one permanent (the validator refuses that
// cast; this only keeps a malformed announcement from being priced as
// a discount). Caller must hold g.mu.
func (g *Game) altSacrificeManaValueLocked(playerID uuid.UUID, card Card, params CastSpellParams) int {
	if len(params.AltCostIDs) != 1 {
		return 0
	}
	alt := g.claimedAltCostLocked(playerID, card, params)
	if alt == nil || !alt.ReducedBySacrificedManaValue {
		return 0
	}
	return g.AltSacrificeManaValueForEffect(params.AltCostIDs[0])
}

// AltSacrificeManaValueForEffect is the mana value emerge reduces its
// cost by when `id` is the permanent sacrificed (CR 702.119a): the
// permanent's mana value as it stands on the battlefield (CR 202.3), so
// a token that isn't a copy is 0, a copy has the copied card's, a
// face-down permanent is 0 and an {X} in its cost counts 0 (CR 202.3e).
// Zero when `id` names no permanent.
//
// Exported for the readers that price a payment before announcing it —
// the bot enumerator's per-payment price and the view's per-candidate
// one — so all three read the number the payment is charged with.
// Caller must hold g.mu.
func (g *Game) AltSacrificeManaValueForEffect(id uuid.UUID) int {
	if g.Battlefield == nil {
		return 0
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return permanentManaValue(&g.Battlefield.Cards[i])
		}
	}
	return 0
}

// AltCostPermanentsForEffect reads `refs` — a PaidCost's or a
// CastProvenance's AltCostObjects — as the permanents they last were (CR
// 608.2h, ADR 0135 §4): each through PermanentForEffect, so the counters
// and anthems it had count, and an object that has come back as a new
// one since is never mistaken for it. A ref with no battlefield record (a
// pitched card from a hand, an object gone by a route that bypasses the
// battlefield exit) is skipped, which errs weaker. Nil for no refs.
//
// Caller must hold g.mu in write mode, like PermanentForEffect.
func (g *Game) AltCostPermanentsForEffect(refs []ObjectRef) []PermanentInfo {
	if len(refs) == 0 {
		return nil
	}
	out := make([]PermanentInfo, 0, len(refs))
	for _, ref := range refs {
		if info, ok := g.PermanentForEffect(ref); ok {
			out = append(out, info)
		}
	}
	return out
}
