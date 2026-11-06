package heuristic

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// lifegain_payoff.go — #4979: a lifelink grant on a creature in combat is
// worth what the bot's own lifegain payoffs pay for the life it gains.
//
// Heliod, Sun-Crowned puts a +1/+1 counter on a creature or enchantment
// its controller controls whenever that player gains life, and Ajani's
// Pridemate one on itself. #4302 priced Heliod's own "{1}{W}: Another
// target creature gains lifelink until end of turn" on the bot's attacker
// by the attack it changes (pumpAttackGain: the life) and CombatGrant,
// so the counter that life makes was worth nothing.
//
// A "whenever you gain life" row now declares `life_gain_payoff`: what it
// pays for each life-gain event, in landfall.go's units. Such a trigger
// fires once per event, not once per point (CR 119.9), and the damage of
// one lifelink source is one event (CR 702.15b; two sources dealing
// damage at once are two, CR 702.15e). So with
// Config.PriceLifeGainPayoffs, a grant of lifelink to a creature of the
// bot's that lacks it, has power to deal damage, and is attacking or
// blocking before damage, adds one event's worth of the bot's payoffs per
// such creature.
//
// Only when the grant passes spareCombatGrant: in combat, and paid with
// mana the rest of the turn does not need. Before attacks the grant waits
// for combat as #4302 has it (the creature may never deal damage), and a
// counter does not outbid the main-phase cast the mana is kept for
// (#4302 decision 4).
//
// Not priced: the attack of a creature that already has lifelink, double
// strike's second event, and the life a spell, an enters trigger or a
// landfall row gains.

// lifeGainPayoff is what one life-gain event of the bot's pays it: the
// sum, over the triggered rows of the permanents it controls, of each
// declared lifegain payoff. Zero when the Config does not price them.
func (st *state) lifeGainPayoff(cfg Config) float64 {
	if !cfg.PriceLifeGainPayoffs || st.view == nil {
		return 0
	}
	var v float64
	cards := st.view.Battlefield.Cards
	for i := range cards {
		src := &cards[i]
		if src.Controller != st.me {
			continue
		}
		for _, r := range src.AbilityRows {
			if r.Kind != "triggered" || r.Purpose == nil || r.Purpose.LifeGainPayoff == nil {
				continue
			}
			l := protocol.LandfallPayoffView(*r.Purpose.LifeGainPayoff)
			v += st.landfallRowValue(cfg, &l)
		}
	}
	return v
}

// lifelinkGrantPayoff is what move m's lifelink grants pay through the
// bot's lifegain payoffs: one life-gain event for each creature of its
// own in combat that m gives lifelink it lacks and that has power to deal
// damage. Zero unless m passes spareCombatGrant, and for a grant the same
// row already gave the same creature this turn (repeatsTemporaryGrant).
func (p *Policy) lifelinkGrantPayoff(st *state, m legal.Move) float64 {
	if !p.cfg.PriceLifeGainPayoffs || (m.Kind != legal.KindActivate && m.Kind != legal.KindCast) {
		return 0
	}
	per := st.lifeGainPayoff(p.cfg)
	// A grant the row already made this turn is still on the stack or
	// already on the creature: its event is counted once.
	if per <= 0 || !p.spareCombatGrant(st, m) || p.repeatsTemporaryGrant(st, m) {
		return 0
	}
	picks, entries := st.temporaryPicks(m, p.cfg.PriceXPumps)
	n := 0
	for i, t := range picks {
		c, e := st.bf[t.ID], entries[i]
		if c == nil || c.Controller != st.me || e.UntilEndOfTurn != targetBenefit || e.Pump == nil {
			continue
		}
		if slices.Contains(e.Pump.Keywords, "lifelink") && !hasKeyword(c, "lifelink") && combatPower(c)+e.Pump.Power > 0 {
			n++
		}
	}
	return float64(n) * per
}
