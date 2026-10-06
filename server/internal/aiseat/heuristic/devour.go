package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// devour.go — the bot's answer to a devour prompt (#2419, CR 702.82a).
//
// "As this creature enters, you may sacrifice any number of creatures.
// It enters with N +1/+1 counters on it for each creature sacrificed
// this way." The prompt shares entry_sacrifice with the sacrifice
// lands (Heart of Yavimaya), where every offered set has the same size
// and the cheapest set wins. Devour differs: the empty answer is legal
// and scores zero, so a set wins only if what it buys beats what it
// eats.
//
// A set is priced as the sum over its creatures, which keeps the answer
// cheap (each enumerated combination scores in one pass) and makes the
// bot eat exactly the creatures worth less than their counters.
//
//	gain  = N * (Power + Toughness)        the counters, on the same
//	                                       scale permanentValue uses
//	      + draw * drawValue               Skullmulcher's cards
//	      + life * Life                    Marrow Chomper's life
//	cost  = fuelValue(creature)            what it is worth on the board
//	      + DevourPermanent                nontoken: a card, not a body
//	      + DevourCommander                the bot's own commander
//
// Not modelled: a creature's ability beyond its stat line and keywords,
// dies triggers of the devoured creature, Doubling Season on the
// counters. An engine the stat line does not show is protected only by
// DevourPermanent.
func (p *Policy) devourValue(st *state, ch *protocol.PendingChoiceView, ids []string) (float64, string) {
	if len(ids) == 0 {
		return 0, "devour nothing"
	}
	perCreature := float64(ch.Devour)*(st.w.Power+st.w.Toughness) +
		float64(ch.DevourLife)*st.w.Life
	if ch.DevourDraw > 0 {
		// A walk over the seat's cards, so only when a card is paid.
		perCreature += float64(ch.DevourDraw) * p.drawValue(st)
	}
	var v float64
	for _, id := range ids {
		v += perCreature - p.fuelValue(st, id)
		if c := st.bf[id]; c != nil {
			if c.IsCommander {
				v -= p.cfg.DevourCommander
			}
			if !c.IsToken {
				v -= p.cfg.DevourPermanent
			}
		}
	}
	return v, "devour: counters against what the creatures are worth"
}
