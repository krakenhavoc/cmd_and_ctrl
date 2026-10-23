package heuristic

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// ignore_effect.go — #1986, CR 116.2d: the bot's price for ignoring a
// static ability by sacrificing a permanent (Damping Engine).
//
// The enumerator offers the ignore to a seat the effect bears on, paying
// with the first permanent of the payment order (a token, then the
// cheapest). The sacrifice is a real cost, so the policy takes the ignore
// only when it buys something it would do this turn: on its own turn, with
// the stack empty, a card in its hand that the effect refuses (the view's
// `cant_cast` names the source) and that it could play now. A land is worth
// a land drop; a spell is worth half what casting it is worth, since the
// permanent given up is not free. With nothing to unblock it is priced
// below passing, so a restricted bot sits tight rather than shedding a
// permanent every priority.

// ignoreEffectDeclined is what an ignore that unblocks nothing is worth.
const ignoreEffectDeclined = -1.0

func (p *Policy) valueOfIgnoreEffect(st *state, sp specialActionParams) (float64, string) {
	src := st.bf[sp.CardID]
	if src == nil || st.seat == nil || !st.myTurn || !st.sorcerySpeed {
		return ignoreEffectDeclined, "ignore declined: not a window to play what it unblocks"
	}
	best, why := 0.0, ""
	landDropLeft := st.seat.LandDropsPerTurn > st.seat.LandsPlayedThisTurn
	for i := range st.seat.Hand.Cards {
		c := &st.seat.Hand.Cards[i]
		if c.CantCast == "" || !strings.Contains(c.CantCast, src.Name) {
			continue
		}
		v := p.unblockedValue(st, c, landDropLeft)
		if v > best {
			best, why = v, c.Name
		}
	}
	if best <= 0 {
		return ignoreEffectDeclined, "ignore declined: nothing in hand to unblock"
	}
	return best, "ignore " + src.Name + " to play " + why
}

// unblockedValue is what playing `c` is worth if the effect stopped
// refusing it, 0 when the bot could not play it this turn anyway.
func (p *Policy) unblockedValue(st *state, c *protocol.CardView, landDropLeft bool) float64 {
	if isLand(c) {
		if landDropLeft {
			return p.cfg.LandValue
		}
		return 0
	}
	mv := manaValue(c.ManaCost, 0)
	if mv > st.myMana {
		return 0
	}
	return 0.5 * p.cfg.SpellPerMana * float64(max(mv, 1))
}
