package heuristic

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// spare_token.go prices a SPARE creature token given up to pay a cost
// (#4991): one more copy of a body the bot already has enough of.
//
// sacrificeCost charges a sacrificed permanent its whole board value,
// and a 1/1 flying Thopter is worth 2.95 there (power, toughness and the
// flying bonus) against a drawn card's Weights.Hand of 1.20. So Sai,
// Master Thopterist's "{1}{U}, Sacrifice two artifacts: Draw a card"
// cost the bot six points for one, and it was never activated, however
// many Thopters Sai had made. That price is right for the last few
// bodies on the board and wrong for the fifth Thopter: what the bot
// loses with a copy of a body it keeps two more of is not the body's
// stats, which the kept copies still bring, but one more creature to
// block and attack with.
//
// With Config.PriceSpareTokens on, a creature token sacrificed to pay a
// cost is spare when, after the payment, the bot still controls at least
// Config.SpareTokenKeep other creature tokens with the same body (name,
// power, toughness and keywords). A spare token costs what tapping it
// for a turn cycle does (tapCreatureCost): a blocker the bot no longer
// has, and before combat on the bot's own turn the attack it gives up,
// never less than the blocker. So Sai cashes in two Thopters for a card
// after combat once it has four, and keeps two.
//
// The switch ships OFF. In #4991's arena (thopter-battle, two seats with
// the price against two without, three seed sets of 48 games) the seats
// that traded Thopters for cards won 20.1% against 29.9%: under these
// weights the two fliers were worth more than the card, as the board
// price said. On the curated mono-blue-artifacts deck (48 head-to-head
// games) the price neither won nor lost measurably, 23.9% against
// 26.1%, so it stays off. `heuristic-sparetokens` seats it for the next
// measurement.
//
// Not spare, whatever the count: a nontoken creature (a card can matter
// in ways its body does not show, as blockOnly says), a commander, a
// token in combat, one with counters (a different body from its copies,
// whatever its stats), and one with something attached to it. A token
// that is dying anyway (sacrifice.go) still costs only the share the
// bot would have kept.

// spareToken reports whether c, one of the permanents a cost sacrifices
// (`paid`), is a spare token.
func (p *Policy) spareToken(st *state, c *protocol.CardView, paid []string) bool {
	if !p.cfg.PriceSpareTokens || p.cfg.SpareTokenKeep <= 0 || !spareCandidate(st, c) {
		return false
	}
	kept := 0
	for i := range st.view.Battlefield.Cards {
		o := &st.view.Battlefield.Cards[i]
		if o.InstanceID == c.InstanceID || slices.Contains(paid, o.InstanceID) || !spareCandidate(st, o) || !sameBody(o, c) {
			continue
		}
		kept++
		if kept >= p.cfg.SpareTokenKeep {
			return true
		}
	}
	return false
}

// spareCandidate reports whether c is a creature token of the bot's
// that could be one of a set of interchangeable copies: no commander,
// not in combat, no counters and nothing attached to it.
func spareCandidate(st *state, c *protocol.CardView) bool {
	if c == nil || !c.IsToken || c.IsCommander || c.Controller != st.me || !isCreature(c) {
		return false
	}
	if c.AttackingTarget != "" || c.BlockingTarget != "" || len(c.BlockingTargets) > 0 || len(c.Counters) > 0 {
		return false
	}
	for i := range st.view.Battlefield.Cards {
		if h := st.attach.hostOf(&st.view.Battlefield.Cards[i]); h != nil && h.InstanceID == c.InstanceID {
			return false
		}
	}
	return true
}

// sameBody reports whether two creature tokens are copies of one body:
// the same name, power, toughness and keywords.
func sameBody(a, b *protocol.CardView) bool {
	if a.Name != b.Name || a.Power != b.Power || a.Toughness != b.Toughness || len(a.Abilities) != len(b.Abilities) {
		return false
	}
	for _, kw := range a.Abilities {
		if !slices.Contains(b.Abilities, kw) {
			return false
		}
	}
	return true
}

// spareTokenCost is what a spare token given up to a cost costs: what
// tapping it would (tapCreatureCost), and never less than the blocker
// it is (tappedBlocker), since a sacrificed token does not untap.
func (p *Policy) spareTokenCost(st *state, c *protocol.CardView) float64 {
	return max(p.tapCreatureCost(st, c), tappedBlocker)
}
