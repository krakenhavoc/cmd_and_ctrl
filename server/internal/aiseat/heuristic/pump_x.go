package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// pump_x.go — #4976: a pump until end of turn whose amount is the
// announced X, priced at that X (ADR 0126's amendment of 2026-11-05 for
// #4976).
//
// Tyvar's Stand ("+X/+X and gains hexproof and indestructible"), Hatred
// ("+X/+0") and Mishra's Command's last bullet ("+X/+0 and gains haste")
// declared a benefit with no amount, because the amount is not printed.
// So the attack the pump adds was never priced: a Tyvar's Stand at X=3
// on an attacker before blocks was worth CombatGrant and its keywords,
// the same as at X=0. The entries now declare PumpPowerIsX and
// PumpToughnessIsX, and with Config.PriceXPumps entryAtX adds the move's
// X to the pump, so temporaryEffectValue prices the +3/+3 the move
// gives.
//
// Hatred's X is paid in life, as an additional cost (CR 601.2f): the
// move's x_value is both the life paid and the +X/+0. That life is not
// in the move's declared cost (legal.MoveCost.Life counts only fixed and
// Phyrexian life), and the enumerator offers the largest X the life
// allows, so pricing the pump at that X without its life would have the
// bot pay twenty life for a swing it does not need. pumpXLifeCost
// charges it, as costValue charges a declared life cost but without the
// LifePayoff proxy: the life buys the pump, which is priced.

// paysXLife reports whether a card's additional cost is "pay X life"
// (Toxic Deluge, Hatred): the wire's demands_x with neither of the other
// X costs that set it (blight X, exile X cards).
func paysXLife(c *protocol.CardView) bool {
	if c == nil || c.AdditionalCost == nil {
		return false
	}
	ac := c.AdditionalCost
	return ac.DemandsX && !ac.BlightX && !ac.ExileCostX
}

// pumpXLifeCost is what a cast whose X is paid in life charges for that
// life when one of its picks declares a pump of X (#4976): the price the
// life costs the bot at its total (LifeCostValue), and a refusal when it
// would leave the bot below LifeFloor. Zero, false with PriceXPumps off,
// for a cast whose X is paid in mana, and for one whose entries declare
// no pump of X (Toxic Deluge keeps its price).
func (p *Policy) pumpXLifeCost(st *state, card *protocol.CardView, cp castParams, targets []targetRef) (float64, bool) {
	if !p.cfg.PriceXPumps || cp.XValue <= 0 || !paysXLife(card) {
		return 0, false
	}
	pumps := false
	for _, t := range targets {
		if declaresPumpX(castEntryFor(card, cp, t)) {
			pumps = true
			break
		}
	}
	if !pumps {
		return 0, false
	}
	life := st.myLife()
	if life-cp.XValue < p.cfg.LifeFloor {
		return 0, true
	}
	return st.w.LifeCostValue(life, cp.XValue), false
}
