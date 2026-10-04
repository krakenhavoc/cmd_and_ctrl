package heuristic

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// ring_bearer.go answers "choose your Ring-bearer" (CR 701.54a, ADR
// 0114 §7). The Ring makes its bearer legendary and hard to block, and
// at higher levels it loots when it attacks and drains when it
// connects, so the right bearer is the creature that attacks hardest:
//
//  1. the power it attacks with;
//  2. then whether it can attack this turn (untapped, not sick);
//  3. then the current Ring-bearer, which costs nothing to keep.
//
// Last of all, below every other candidate, is a creature that the
// Ring's "is legendary" would put into the legend rule (CR 704.5j)
// against a same-named legendary permanent the seat already controls:
// choosing it would cost the seat one of the two.

// choiceRingBearer is the wire kind (game.PendingChoiceRingBearer).
const choiceRingBearer = "ring_bearer"

// ringBearerValue prices one candidate: higher is a better
// Ring-bearer.
func (st *state) ringBearerValue(c *protocol.CardView) float64 {
	if c == nil {
		return 0
	}
	if st.legendRuleVictim(c) {
		return -1000
	}
	v := float64(c.Power) * 10
	if !c.Tapped && !c.SummoningSick {
		v += 5
	}
	if c.RingBearer {
		v++
	}
	return v
}

// legendRuleVictim reports whether making `c` legendary would put it
// into the legend rule: it is not legendary now, and the seat controls
// another legendary permanent with the same name.
func (st *state) legendRuleVictim(c *protocol.CardView) bool {
	if isLegendaryLine(c.TypeLine) {
		return false
	}
	for id, o := range st.bf {
		if id == c.InstanceID || o.Controller != c.Controller {
			continue
		}
		if o.Name == c.Name && isLegendaryLine(o.TypeLine) {
			return true
		}
	}
	return false
}

func isLegendaryLine(typeLine string) bool {
	for _, w := range strings.Fields(typeLine) {
		if strings.EqualFold(w, "Legendary") {
			return true
		}
	}
	return false
}

// RingBearerOrder implements aiseat.RingBearerOrderer: the same price
// the decision uses, so what survives the enumerator's cap is what the
// decision would have picked anyway.
func (p *Policy) RingBearerOrder(in aiseat.Input) legal.RingBearerOrder {
	p.mu.Lock()
	st := p.newState(in)
	p.mu.Unlock()
	return func(c legal.TargetCandidate) float64 {
		return st.ringBearerValue(st.bf[c.ID.String()])
	}
}

// compile-time proof that the heuristic really is a RingBearerOrderer.
var _ aiseat.RingBearerOrderer = (*Policy)(nil)
