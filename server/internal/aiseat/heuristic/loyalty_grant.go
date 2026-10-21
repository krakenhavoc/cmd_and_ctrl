package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// loyalty_grant.go — #4191: a row that gives each of the bot's
// planeswalkers one more loyalty activation this turn. The Chain Veil
// prints it: "{4}, {T}: For each planeswalker you control, you may
// activate one of its loyalty abilities once this turn as though none
// of its loyalty abilities have been activated this turn." It declares
// Purpose.ExtraLoyaltyActivations.
//
// What the row buys is used only in the bot's own main phase with an
// empty stack (a loyalty ability's timing, CR 606.3), and it ends with
// the turn. So it is worth ActivateBase for each grant and each
// planeswalker of the bot's that has a loyalty row, when the bot is in
// such a window now, and below passing anywhere else: flat-priced, the
// bot would spend {4} on it in an opponent's end step, or with no
// planeswalker at all. After it resolves the enumerator offers the
// walkers' loyalty rows again (Game.LoyaltySpentLocked), and they are
// priced as they always are.

// loyaltyGrantValue prices an activation of a row that declares
// ExtraLoyaltyActivations. ok is false for every other row.
func (st *state) loyaltyGrantValue(cfg Config, src *protocol.CardView, index int) (v float64, ok bool) {
	pv := rowPurpose(src, index)
	if pv == nil || pv.ExtraLoyaltyActivations <= 0 {
		return 0, false
	}
	if !st.sorcerySpeed {
		return idleEquipMove, true
	}
	walkers := 0
	for _, c := range st.bf {
		if c.Controller == st.me && isType(c, "planeswalker") && hasLoyaltyRow(c) {
			walkers++
		}
	}
	if walkers == 0 {
		return idleEquipMove, true
	}
	return cfg.ActivateBase * float64(walkers*pv.ExtraLoyaltyActivations), true
}

// hasLoyaltyRow reports whether c has an activated row with a loyalty
// cost (CR 606.2), printed or granted.
func hasLoyaltyRow(c *protocol.CardView) bool {
	for i := range c.ActivatedAbilities {
		if r := &c.ActivatedAbilities[i]; r.LoyaltyCost != nil || r.LoyaltyCostX {
			return true
		}
	}
	return false
}
