package game

import (
	"fmt"

	"github.com/google/uuid"
)

// component_cost_modifier.go — #6206: a cost change aimed at PART of a
// cost. "Buyback costs cost {2} less" (Memory Crystal) is not a
// reduction of the spells it touches: it is a reduction of one
// component of their cost, the buyback additional cost (CR 702.27a),
// and it must leave every other part of the total alone.
//
// # Why the whole-cost pass cannot say it
//
// CR 601.2f totals a cast as "the mana cost or alternative cost, plus
// all additional costs and cost increases, minus all cost reductions",
// and the engine's one pricer (printedCostLocked) folds the announced
// additional costs' mana into one ParsedCost before the CR 601.2f
// modifier pass runs. By then the buyback {3} and the spell's own {U}
// are the same number, and an ordinary {2} CostReduction aimed at that
// number gets both halves of the rule wrong:
//
//   - it discounts a cast that paid no buyback at all, because the
//     component it names is not part of the question the whole-cost
//     pass asks; and
//   - it spends itself on whatever generic mana it finds first (CR
//     118.7a), so a {2}{U} spell with buyback {3} would come out at
//     {3}{U} instead of {2}{U}{3-2}.
//
// So the component is priced ON ITS OWN, by its own CR 601.2f pass over
// the modifiers aimed at it, and what that pass leaves is the mana that
// joins the total (AdditionalCostMana). Three rules fall out of that and
// are worth stating:
//
//   - The floor is the component's, not the cast's (CR 601.2f, 118.7):
//     {2} less off a buyback {1} leaves {0} and takes nothing off the
//     spell's own mana. Reducing part of a cost can never reduce the
//     rest of it.
//   - Order does not matter. A component reduction never touches the
//     spell's own mana and a whole-cost increase never adds to a
//     component, so "increases before reductions" is unobservable
//     between the two passes — a Thalia on a bought-back spell adds its
//     {1} to the total either way.
//   - The spell's mana cost is untouched (CR 118.8d): an additional cost
//     is not the mana cost, so nothing here moves what a card "costs" to
//     anything that reads a mana cost.
//
// # What reaches a component
//
// The board's modifiers, through the one gatherer every pass uses
// (activeCostModifiersLocked): a permanent's static, an eminence
// commander's, a plane's, an emblem's, a scoped tax or reduction. Keyed
// by CostModifier.PricesOptionalCost against
// CostQuery.PricingOptionalCost, so the partition is total in both
// directions — a Sphere of Resistance never taxes a buyback cost, and
// Memory Crystal never discounts a spell. A spell's own self modifier
// and a "next spell you cast" promise are written against the spell's
// own cost and never reach a component (CR 118.8d).
//
// Only an OPTIONAL additional cost is priced this way: every printed
// clause of this family names one by its keyword ("buyback costs",
// "kicker costs"). The mandatory slot has no key, and a cost with no
// mana half — "Buyback—Sacrifice a land" (Constant Mists),
// "Buyback—Pay 4 life" (Slaughter) — has nothing a mana reduction can
// take (CR 118.7a), so both are left exactly as printed.

// PricedComponents is the cost each payment of one announcement's
// additional costs charges after the modifiers aimed AT THAT COMPONENT
// have been applied — one entry per entry of the castCostPayments plan,
// in plan order, and nil for every payment nothing reprices (which is
// every payment on every board without such a modifier).
//
// Indexed by position rather than keyed by the cost's Key, because a
// card may print two costs with the same key (CR 702.33b's two kickers)
// and a multikicker paid three times is three payments of one key: the
// plan's order is the one thing that tells them apart, and it is the
// order the payer and the validator already walk.
//
// Plain data rather than a hook, so AdditionalCostMana stays a pure
// function and the ADR 0041 closure ratchet gains no route from Game.
type PricedComponents []*ParsedCost

// at is the cost payment `i` charges, and false when the component was
// not repriced. Nil-safe, so every caller can pass the zero value.
func (p PricedComponents) at(i int) (ParsedCost, bool) {
	if i < 0 || i >= len(p) || p[i] == nil {
		return ParsedCost{}, false
	}
	return *p[i], true
}

// pricedComponentsLocked prices every payment of this announcement's
// additional costs that a component cost modifier is aimed at. Nil for
// nearly every cast: the plan must announce an optional cost with a
// key and a mana half, and something on the board must price that key.
//
// Called by the ONE pricer (printedCostLocked), so the payment, the
// auto-tap preview, the view's badge and the bot enumerator all read
// the same component price (#544).
//
// Caller must hold g.mu.
func (g *Game) pricedComponentsLocked(p *Player, card Card, params CastSpellParams, mandatory *AdditionalCost, optional []AdditionalCost) (PricedComponents, error) {
	if g == nil || p == nil {
		return nil, nil
	}
	plan := castCostPayments(mandatory, optional, params.OptionalCosts)
	var out PricedComponents
	for i, pay := range plan {
		// The mandatory slot (index -1) carries no keyword and no key;
		// a cost with no mana half has nothing a mana modifier can take.
		if pay.index < 0 || pay.cost.Key == "" || pay.cost.ManaCost == "" {
			continue
		}
		q := g.componentCostQueryLocked(p.ID, card, params, pay.cost.Key)
		mods := g.activeCostModifiersLocked(q)
		if len(mods) == 0 {
			continue
		}
		before, err := ParseCost(pay.cost.ManaCost)
		if err != nil {
			return nil, fmt.Errorf("%w for %s: %w", ErrUnparseableCost, pay.cost.Label, err)
		}
		// An {X} in a component is no honest price: the reduction a
		// printed generic cannot absorb is kept on ParsedCost.XReduced
		// for the cost it was measured against (CR 107.3a, 118.7a), and
		// the component's X is not the spell's. No printed optional cost
		// with a keyword carries {X} with a modifier aimed at it, so
		// refuse rather than guess — the file header's rule, and
		// amountFor's answer to a cost it cannot price.
		if before.XSlots > 0 {
			return nil, fmt.Errorf("%w: %q prices %s, whose cost carries {X}",
				ErrCostModifier, mods[0].modifier.Label, pay.cost.Label)
		}
		after, err := g.applyCostModifiersLocked(before, q)
		if err != nil {
			return nil, err
		}
		if out == nil {
			out = make(PricedComponents, len(plan))
		}
		out[i] = &after
	}
	return out, nil
}

// componentCostQueryLocked builds the CostQuery for one component of
// one cast. One builder, for abilityCostQueryLocked's reason: the
// payment path and everything that quotes it ask the board the same
// shaped question about the same component (#544).
//
// It carries only what a component modifier may read: the spell, its
// caster, the zone it is cast from, the announced X, the announced
// targets and the announced optional costs. Every field the whole-cost
// pass fills from the announcement's OTHER prices — the alternative
// cost's reductions, a permission's "costs {2} less", offering's and
// emerge's — is deliberately zero here: each of those is a change to the
// cost the spell is being cast for, not to an additional cost paid
// beside it (CR 118.8d), and applying one twice would discount the cast
// twice.
//
// Caller must hold g.mu.
func (g *Game) componentCostQueryLocked(caster uuid.UUID, card Card, params CastSpellParams, key string) CostQuery {
	fromZone, _ := castZoneFromWire(params.FromZone)
	return CostQuery{
		Game:                g,
		Card:                card,
		Controller:          caster,
		FromZone:            fromZone,
		XValue:              params.XValue,
		Targets:             params.Targets,
		OptionalCosts:       params.OptionalCosts,
		PricingOptionalCost: key,
	}
}
