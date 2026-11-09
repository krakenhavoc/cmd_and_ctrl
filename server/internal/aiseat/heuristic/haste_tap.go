package heuristic

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// haste_tap.go — #4324 (#4312's item 4): haste granted until end of turn
// is also worth the {T} ability it lets a summoning-sick creature use.
//
// A creature that has not been under its controller's control since their
// most recent turn began can neither attack nor pay a {T} cost (CR 302.6),
// and haste lifts both (CR 702.10b, 702.10c). temporaryBenefitValue prices
// a declared haste grant only by the attack it adds (pumpAttackGain), so
// Expedite or Hanweir Battlements aimed at a fresh creature whose value is
// its "{T}: draw a card" was worth nothing, and the bot never made it.
//
// Everything this needs is already on the wire: the entry's declared
// keyword, the creature's summoning_sick flag and its ability rows with
// their tap_cost, mana cost and purpose. The grant is now worth the best
// such row the bot could activate once the grant resolves, less the
// tapped blocker the activation pays, when that beats what the grant was
// already priced at (the attack, or a save). The row must cost only mana
// and the tap, and the bot must still have that mana after paying for
// the grant (and after tapping the grant's own source, when that is a
// mana source, as Hanweir Battlements is). A row is priced as the activation is priced: its declared
// purpose when it has one, ActivateBase when it does not. Rows that wait
// for something to answer (ADR 0142), rows that sacrifice their source,
// and mana abilities are left out; a hasted mana creature's one extra
// mana this turn is not priced.

// hasteTapGain is what move m's haste grants add beyond the price
// pricedTargetsValue already gave their picks: the largest such excess
// over m's picks, never below zero. Zero unless Config.PriceHasteTapRows
// and PriceTemporaryEffects are on.
func (p *Policy) hasteTapGain(st *state, m legal.Move) float64 {
	if !p.cfg.PriceHasteTapRows || !p.cfg.PriceTemporaryEffects {
		return 0
	}
	picks, entries := st.temporaryPicks(m, p.cfg.PriceXPumps)
	if len(picks) == 0 {
		return 0
	}
	left := st.myMana - st.manaSpentBy(m)
	if m.Kind == legal.KindActivate {
		// A row that taps its own mana source (Hanweir Battlements'
		// "{R}, {T}") also takes that source's mana away.
		cp := decode[activateParams](m.Params)
		if src := st.bf[cp.SourceCardID]; src != nil {
			if row := activatedRow(src, cp.AbilityIndex); row != nil && row.TapCost {
				left -= repeatableMana(src)
			}
		}
	}
	best := 0.0
	for i, t := range picks {
		e := entries[i]
		if e.UntilEndOfTurn != targetBenefit || e.Pump == nil || !slices.Contains(e.Pump.Keywords, "haste") {
			continue
		}
		c := st.bf[t.ID]
		if !hasteUnlocksTap(st, c) {
			continue
		}
		row, ok := p.bestTapRow(st, c, left)
		if !ok {
			continue
		}
		priced, _ := p.temporaryEffectValue(st, t, e)
		best = max(best, row-priced)
	}
	return best
}

// hasteUnlocksTap reports whether haste would let the bot pay a {T} cost
// of its creature c that it cannot pay now: c is untapped, summoning-sick
// and has no haste, and nothing already lets it activate as though it had
// haste (Thousand-Year Elixir, CR 609.4).
func hasteUnlocksTap(st *state, c *protocol.CardView) bool {
	return c != nil && c.Controller == st.me && isCreature(c) && c.SummoningSick &&
		!c.ActivatesAsThoughHaste && !c.Tapped && !hasKeyword(c, "haste")
}

// bestTapRow is the best price, less the tapped blocker, among c's {T}
// rows the bot could activate with `left` mana once c has haste. ok is
// false when there is none worth more than nothing.
func (p *Policy) bestTapRow(st *state, c *protocol.CardView, left int) (float64, bool) {
	best, ok := 0.0, false
	for i := range c.ActivatedAbilities {
		row := &c.ActivatedAbilities[i]
		if !tapRowUnlockable(st, row, left) {
			continue
		}
		v := p.cfg.ActivateBase
		var ps purposeSet
		ps.add(row.Purpose)
		if p.purposePriced(ps) {
			v = p.purposeValue(st, ps, 0, nil, false)
		}
		v -= p.tapCreatureCost(st, c)
		if v > best {
			best, ok = v, true
		}
	}
	return best, ok
}

// tapRowUnlockable reports whether row is a {T} ability whose only other
// cost is mana the bot has `left`, that the bot could activate in this
// window once its source has haste, and that is not a defensive answer
// waiting for the stack.
func tapRowUnlockable(st *state, row *protocol.ActivatedAbilityView, left int) bool {
	if !row.TapCost || row.SacrificeSelf || row.ConditionUnmet || row.Exhausted || row.CantActivate != "" ||
		row.LoyaltyCost != nil || row.LifeCost > 0 || row.EnergyCost > 0 || row.EnergyCostX || row.MillCost > 0 ||
		row.SacrificeLabel != "" || row.CrewCost > 0 || row.Saddle || row.CounterCostN > 0 || row.CounterCostVariable ||
		row.DiscardSelf || row.ExileSelf || row.ReturnSelf || row.BottomSelf || row.ExileGrantor || row.TapGrantor || row.Exert {
		return false
	}
	if row.SorcerySpeed && !st.sorcerySpeed {
		return false
	}
	if pv := row.Purpose; pv != nil && pv.Answers != nil && len(*pv.Answers) > 0 {
		return false
	}
	cost := row.ManaCost
	if row.ChargedManaCost != nil {
		cost = *row.ChargedManaCost
	}
	if costHasX(cost) {
		return false
	}
	return manaValue(cost, 0) <= left
}
