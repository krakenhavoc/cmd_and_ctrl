package game

import "github.com/google/uuid"

// loyalty_cost_add.go — "Planeswalkers' loyalty abilities you activate
// cost an additional [+1] to activate." (Carth the Lion, #6148).
//
// CR 606.4: "The cost to activate a loyalty ability of a permanent is to
// put on or remove from that permanent a certain number of loyalty
// counters, as shown by the loyalty symbol in the ability's cost. This
// cost may be modified by other effects." CR 606.5: "If the total cost to
// activate a loyalty ability contains multiple costs to add or remove
// loyalty counters, those costs are combined into a single cost to add or
// remove loyalty counters, as appropriate." So a −2 under Carth is a −1, a
// +1 is a +2, a 0 is a +1, and CR 606.6 judges the combined cost: a −3 is
// activatable at 2 loyalty because it is a −2.
//
// The addition is catalog data on the permanent that prints it
// (CardDef.LoyaltyCostAdd), summed over the activator's permanents on the
// battlefield. LoyaltyAdjustedCostForEffect folds it into the ability's
// cost once, so every reader of the cost — the activation's CR 606.6 check
// and payment, the enumerator and the view — sees the combined number.
//
// The mana half of #6148 ("Loyalty abilities of planeswalkers your
// opponents control cost {1} more to activate", Eidolon of Obstruction)
// is an ordinary activation cost increase that reads
// AbilityCostSubject.Loyalty.

// CatalogLoyaltyCostAdd is the catalog hook for a permanent whose static
// ability adds loyalty to the loyalty abilities its controller activates;
// carddef.go's init wires it. Nil (no catalog) means none does.
var CatalogLoyaltyCostAdd func(oracleID string) int

// loyaltyCostAddLocked is the loyalty every loyalty ability `activator`
// activates costs in addition: the sum of CardDef.LoyaltyCostAdd over the
// permanents they control (a phased-out one is not on g.Battlefield, CR
// 702.26b; one that lost its abilities, CR 613.1f, adds nothing). Caller must hold g.mu.
func (g *Game) loyaltyCostAddLocked(activator uuid.UUID) int {
	if g == nil || g.Battlefield == nil || CatalogLoyaltyCostAdd == nil {
		return 0
	}
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != activator {
			continue
		}
		if key := catalogAbilityKeyOf(c); key != "" {
			n += CatalogLoyaltyCostAdd(key)
		}
	}
	return n
}

// LoyaltyAdjustedCostForEffect is `cost` as `activator` pays it: for a
// loyalty ability, the printed loyalty symbol combined with every addition
// on the board (CR 606.5) into one Loyalty; every other cost unchanged. The
// returned cost's Loyalty is a fresh pointer, never the catalog's. Caller
// must hold g.mu.
func (g *Game) LoyaltyAdjustedCostForEffect(activator uuid.UUID, cost AbilityCost) AbilityCost {
	if cost.Loyalty == nil {
		return cost
	}
	add := g.loyaltyCostAddLocked(activator)
	if add == 0 {
		return cost
	}
	n := *cost.Loyalty + add
	cost.Loyalty = &n
	return cost
}
