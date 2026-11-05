package heuristic

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// spare_tokens.go is #4991: a pile of identical creature tokens past the
// first few is fodder for a sacrifice cost.
//
// sacrificeCost charged every permanent its whole board value, and a 1/1
// flying Thopter is worth 2.95 there (power, toughness and the flying
// bonus). Sai, Master Thopterist's "{1}{U}, Sacrifice two artifacts: Draw a
// card" bought a card (Weights.Hand, 1.20) for 5.90 of Thopters, so the bot
// never drew with it, however many Thopters it had. Each copy of a token is
// priced as if it were the only one, which is what a board of five Thopters
// is not: an opponent's flier, a sweeper or a lifegain answers them alike,
// and the fifth one adds less than the first.
//
// Under Config.PriceSpareTokens a creature token sacrificed to pay a cost
// is SPARE when, after the payment, the bot still controls at least
// Config.SpareTokenKeep creature tokens with its name. A spare token costs
// Config.SpareTokenShare of its board value. Read off the view: a token's
// name, `is_token`, its counters and what is attached to it (ADR 0033 §3).
//
// Not spare, whatever the pile: a token with counters on it or something
// attached to it (it is no longer like the others), a token another seat
// controls, and, before combat in the bot's own turn, a token that could
// attack, so the Thopters swing first and are cashed in after combat. The kept tokens are the bot's army and stay at full price,
// so an outlet never eats the pile below SpareTokenKeep. The chance the
// token was dying anyway (sacrifice.go) still applies on top.

// spareTokens reports which of the permanents `ids`, sacrificed together to
// pay one cost, are spare tokens (see above). Nil when none is, or when
// Config.PriceSpareTokens is off.
func (p *Policy) spareTokens(st *state, ids []string) map[string]bool {
	if !p.cfg.PriceSpareTokens || p.cfg.SpareTokenKeep <= 0 || len(ids) == 0 {
		return nil
	}
	var spare map[string]bool
	left := map[string]int{}
	for _, id := range ids {
		c := st.bf[id]
		if !st.plainOwnToken(c) {
			continue
		}
		// Before combat on the bot's own turn a token that could attack
		// still has that attack to make: it is spent after combat.
		if st.myTurn && st.step == "precombat_main" && couldAttack(c) {
			continue
		}
		n, ok := left[c.Name]
		if !ok {
			n = st.tokensNamedLeft(c.Name, ids)
			left[c.Name] = n
		}
		if n < p.cfg.SpareTokenKeep {
			continue
		}
		if spare == nil {
			spare = map[string]bool{}
		}
		spare[id] = true
	}
	return spare
}

// plainOwnToken reports whether c is a creature token the bot controls
// with no counters and nothing attached to it: one of a pile of identical
// copies.
func (st *state) plainOwnToken(c *protocol.CardView) bool {
	if c == nil || !c.IsToken || c.Controller != st.me || !isCreature(c) || c.Name == "" {
		return false
	}
	for _, n := range c.Counters {
		if n != 0 {
			return false
		}
	}
	for _, o := range st.bf {
		if h := st.attach.hostOf(o); h != nil && h.InstanceID == c.InstanceID {
			return false
		}
	}
	return true
}

// tokensNamedLeft counts the creature tokens named `name` the bot controls
// that are not among the permanents `sacrificed`: the pile a payment
// leaves behind.
func (st *state) tokensNamedLeft(name string, sacrificed []string) int {
	n := 0
	for _, c := range st.bf {
		if c.IsToken && c.Controller == st.me && c.Name == name && isCreature(c) && !slices.Contains(sacrificed, c.InstanceID) {
			n++
		}
	}
	return n
}
