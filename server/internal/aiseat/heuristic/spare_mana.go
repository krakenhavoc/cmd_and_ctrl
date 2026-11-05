package heuristic

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// spare_mana.go prices a SPARE mana rock given up to pay a cost (#5085):
// a source whose mana the bot no longer needs to cast what it holds.
//
// A rock on the battlefield is priced by the mana it makes (ADR 0126 §2,
// Weights.ManaSource: 1.00 untapped), and a sacrifice costs a permanent's
// board value, so Mind Stone's "{1}, {T}, Sacrifice Mind Stone: Draw a
// card" was a card (Weights.Hand, 1.20) less the rock (1.00): 0.20,
// under PassThreshold (0.25) in every window. In #4991's thopter-battle
// arena it was offered about 6,000 times in 16 games and never taken.
// That price is right early, when the rock's mana is what stands between
// the bot and its hand, and wrong late: the seventh mana source of a bot
// holding nothing it cannot already cast is worth less than a card.
//
// With Config.PriceSpareManaSources on, a mana rock sacrificed to pay a
// cost is spare when the mana the bot's other sources make, leaving out
// every permanent the cost gives up, still covers both what its hand and
// commander want (manaWant, ramp's deficit) and Config.SpareManaKeep. A
// spare rock costs Config.SpareManaSourceCost (0.50), what is left of a
// source the bot has no cast for: the odd activation or commander tax it
// might still pay. So Mind Stone nets 0.70 and is cashed in once the bot
// has six sources and nothing in hand it cannot cast, and kept while a
// card in hand still wants its mana.
//
// A rock is a noncreature, nonland permanent with a repeatable mana
// ability (repeatableMana). A land is not one: a land is priced on its
// own terms and its sacrifice is landfall.go's business. A mana creature
// is not one either: its body is a blocker whatever its mana is worth.
// A commander is never spare.

// spareManaSource reports whether c, one of the permanents a cost gives
// up (`paid`, which may leave c out when c is the row's own source), is
// a spare mana rock.
func (p *Policy) spareManaSource(st *state, c *protocol.CardView, paid []string) bool {
	if !p.cfg.PriceSpareManaSources || c == nil || c.Controller != st.me || c.IsCommander ||
		isLand(c) || isCreature(c) || repeatableMana(c) <= 0 {
		return false
	}
	others := 0
	for _, o := range st.bf {
		if o.Controller != st.me || o.InstanceID == c.InstanceID || slices.Contains(paid, o.InstanceID) {
			continue
		}
		others += repeatableMana(o)
	}
	return others >= max(p.manaWant(st, nil), p.cfg.SpareManaKeep)
}

// spareManaCost is what a mana rock given up to a cost costs: its whole
// value, or SpareManaSourceCost when it is spare and that is less.
func (p *Policy) spareManaCost(st *state, c *protocol.CardView, paid []string, full float64) float64 {
	if p.spareManaSource(st, c, paid) {
		return min(full, p.cfg.SpareManaSourceCost)
	}
	return full
}
