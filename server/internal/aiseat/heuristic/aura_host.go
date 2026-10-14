package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// aura_host.go prices an answer to CR 303.4f's aura_host prompt (#3158):
// what an Aura put onto the battlefield without being cast enchants as it
// enters. One candidate per answer, a card or a player.
//
// Three cases, read off where the candidate is:
//
//   - A CARD IN A GRAVEYARD. The only Auras that may enchant one are the
//     reanimation family (Animate Dead, Dance of the Dead, ADR 0154): the
//     Aura's enters trigger puts that card onto the battlefield under the
//     bot's control, whoever owned it. So it is worth what the card is
//     worth on the battlefield — its creature value — and the bot takes
//     the best creature in any graveyard. cardValue's "can't cast it for
//     a while" discount does not apply: nothing is cast.
//   - A PERMANENT OR A PLAYER. The prompt carries no oracle text, so the
//     policy reads the Aura the way it reads an Aura SPELL's target
//     (targetsValue): an opponent's permanent is priced as answered, and
//     the bot's own as the small OwnPermanentTarget. A harmful Aura
//     (Pacifism, a Curse) therefore goes on an opponent's best creature or
//     onto the opponent most worth hurting. Among the bot's own permanents
//     — the only candidates of an "enchant creature you control" Aura —
//     the creature worth most wins the tie.
//
// An Aura that declares what it does to its host (Purpose.Aura, #3309)
// is priced by it, on a battlefield permanent: a buff goes on the bot's
// own and costs its value on an opponent's, a curse costs the value of
// the bot's own permanent. An Aura that declares nothing keeps the
// reading above, so an undeclared buff is still priced as a harmful one.

// The wire names of game.AuraPurpose.
const (
	auraBenefit = "benefit"
	auraHarm    = "harm"
)

// auraPurposeOf is the Aura purpose a card declares, "" for none.
func auraPurposeOf(c *protocol.CardView) string {
	if c == nil || c.Purpose == nil {
		return ""
	}
	return c.Purpose.Aura
}

// auraHostPrice prices a battlefield permanent as the host of an Aura
// that declares purpose. False for anything it does not decide (a
// player, a graveyard card, an opponent's host for a curse), which keeps
// the ordinary target price.
func (st *state) auraHostPrice(cfg Config, purpose string, t targetRef) (float64, bool) {
	if t.Kind != "card" || (purpose != auraBenefit && purpose != auraHarm) {
		return 0, false
	}
	c := st.bf[t.ID]
	if c == nil {
		return 0, false
	}
	mine := c.Controller == st.me
	if mine == (purpose == auraHarm) {
		// A buff on an opponent's permanent, a curse on our own.
		return -st.hostWorth(cfg, c), true
	}
	if purpose == auraBenefit {
		return st.hostWorth(cfg, c), true
	}
	return 0, false
}

// hostWorth is what a permanent is worth as the host of an Aura, priced
// the way removing it is (cardTargetValue) without the leader premium.
func (st *state) hostWorth(cfg Config, c *protocol.CardView) float64 {
	base := st.permanentValue(c)
	if isCreature(c) {
		base = st.w.CreatureValue(c)
	}
	return base * cfg.RemovalConfidence
}

// auraTargetsValue prices the targets of an Aura spell that declares
// purpose; a target auraHostPrice does not decide keeps targetsValue's.
func (st *state) auraTargetsValue(cfg Config, purpose string, targets []targetRef) float64 {
	var v float64
	for _, t := range targets {
		if hv, ok := st.auraHostPrice(cfg, purpose, t); ok {
			v += hv
			continue
		}
		v += st.targetsValue(cfg, []targetRef{t})
	}
	return v
}

// auraHostValue prices one aura_host answer. purpose is the prompt's
// aura_purpose, "" when the Aura declares none.
func (st *state) auraHostValue(cfg Config, purpose string, t *targetRef) (float64, string) {
	if t == nil {
		return 0, "enchant nothing"
	}
	if cfg.PriceAuraPurpose {
		if v, ok := st.auraHostPrice(cfg, purpose, *t); ok {
			return v, "enchant the host that suits the Aura"
		}
	}
	if t.Kind == "card" {
		if c := st.graveyard[t.ID]; c != nil {
			if isCreature(c) {
				return st.w.CreatureValue(c), "reanimate the best creature"
			}
			return st.w.permanentValue(c), "enchant the best card in a graveyard"
		}
		if c := st.bf[t.ID]; c != nil && c.Controller == st.me {
			v := cfg.OwnPermanentTarget
			if isCreature(c) {
				v += 0.001 * st.w.CreatureValue(c)
			}
			return v, "enchant our best permanent"
		}
	}
	return st.targetsValue(cfg, []targetRef{*t}), "enchant the best host"
}
