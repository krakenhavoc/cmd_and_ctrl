package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// bestow.go prices a bestowed cast (ADR 0141, #2862, CR 702.103).
//
// The wire carries no "enchanted creature gets" amount, and every bestow
// card's Aura half gives its host about what its own body is: Boon
// Satyr's +4/+2 on a 4/2, Celestial Archon's +4/+4 and keywords on a
// 4/4. So the card's value as a permanent is already the right size for
// the cast either way (resolvedValueFor prices it). What differs is:
//
//   - the pump lands on a creature that may be able to attack now, where
//     the creature cast arrives summoning sick, so a host that is not
//     sick gets back the SickCreature discount;
//   - the Aura survives its host as a creature (CR 702.103f), a second
//     body the creature cast never has, worth BestowShare of the card;
//   - on an opponent's creature it is their bonus, priced as handing them
//     the card twice over, so the bot never bestows onto one.
//
// The bestowed cast costs more mana, and the turn plan weighs that.

// bestowKey is game.BestowKey on the wire: the offer's stable key. The
// policy may not import internal/game.
const bestowKey = "bestow"

// bestowTargetsValue is what a bestowed cast's target is worth: the
// pricing above, in place of targetsValue's, which reads an opposing
// creature as removed. Zero when BestowShare is off; the caller then
// keeps the ordinary target price.
func (p *Policy) bestowTargetsValue(st *state, card *protocol.CardView, targets []targetRef) float64 {
	if card == nil {
		return 0
	}
	body := st.w.permanentValue(card)
	var v float64
	for _, t := range targets {
		host := st.bf[t.ID]
		if host == nil {
			continue
		}
		if host.Controller != st.me {
			v -= 2 * body
			continue
		}
		v += p.cfg.BestowShare * body
		if !host.SummoningSick {
			v += body * (1 - st.w.SickCreature)
		}
	}
	return v
}

// isBestowCast reports whether this cast claims the bestow offer and
// this Config prices it.
func (p *Policy) isBestowCast(cp castParams) bool {
	return p.cfg.BestowShare > 0 && cp.AlternativeCost == bestowKey
}
