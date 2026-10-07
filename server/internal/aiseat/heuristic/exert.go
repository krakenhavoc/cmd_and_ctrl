package heuristic

import (
	"fmt"
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// exert.go prices an exert (ADR 0130 §9, owner decision 4, and the
// amendment of 2026-10-07).
//
// The enumerator offers every attack by an exert creature twice: the
// plain declaration and its twin with `exert: true` (§6). The bot takes
// the twin when it is worth more than the plain attack:
//
//	twin = attackValue(the attacker as the linked row pumps it)
//	     + gain − cost
//
//   - THE GAIN is what the exert's rows declare: every triggered row on
//     the attacker stamped `exert: "linked"` (its "when you do"), and
//     every row stamped `exert: "payoff"` on a permanent the bot
//     controls ("whenever you exert a creature"). A `pump` changes the
//     attacker the attack is evaluated with; the other fields are priced
//     here (exertGain).
//   - THE COST is zero when the exert costs nothing (exertFree): the
//     creature has vigilance, it won't untap during the bot's next untap
//     step anyway, or the attack is a lethal push. Otherwise it is what
//     the creature would be worth next turn on attack, plus what it
//     would be worth as a blocker across the opponents' turns, since it
//     stays tapped through them (exertCost).
//
// Off (Config.PriceExert false, the zero value and BaselineConfig) the
// twin is never taken, which is how the bot played before this file.

// exertMove reports whether m is the twin attack move that exerts its
// attacker.
func exertMove(m legal.Move) bool {
	return m.Kind == legal.KindAttack && decode[attackParams](m.Params).Exert
}

// exertRows are the rows an exert of `atk` would trigger for the bot:
// atk's own linked rows, and every payoff row on a permanent the bot
// controls.
func (st *state) exertRows(atk *protocol.CardView) []*protocol.PurposeView {
	var out []*protocol.PurposeView
	for i := range atk.AbilityRows {
		if r := &atk.AbilityRows[i]; r.Exert == "linked" && r.Purpose != nil {
			out = append(out, r.Purpose)
		}
	}
	for i := range st.view.Battlefield.Cards {
		c := &st.view.Battlefield.Cards[i]
		if c.Controller != st.me {
			continue
		}
		for j := range c.AbilityRows {
			if r := &c.AbilityRows[j]; r.Exert == "payoff" && r.Purpose != nil {
				out = append(out, r.Purpose)
			}
		}
	}
	return out
}

// pumped is the attacker as the exert's linked rows leave it for this
// combat: their pumps added, and a creature none of whose combat damage
// can be dealt read as indestructible, which is what that shield does to
// every exchange attackValue models (kills, deathtouch included).
func pumped(atk *protocol.CardView, rows []*protocol.PurposeView) *protocol.CardView {
	cp := *atk
	cp.Abilities = slices.Clone(atk.Abilities)
	for _, p := range rows {
		if p.Pump != nil {
			cp.Power += p.Pump.Power
			cp.Toughness += p.Pump.Toughness
			cp.Abilities = append(cp.Abilities, p.Pump.Keywords...)
		}
		if p.PreventCombatDamageToSelf {
			cp.Abilities = append(cp.Abilities, "indestructible")
		}
	}
	return &cp
}

// exertGain prices what the exert's rows do besides changing the
// attacker: the §6 amounts the purpose pricer already knows, damage to
// the best creature an opponent controls that it would kill (nothing
// when there is none: a trigger with no legal target is removed, ruling
// 5), damage to each opponent, life, and each additional combat phase,
// priced as the bot's other untapped creatures' attacks in it.
func (p *Policy) exertGain(st *state, atk *protocol.CardView, rows []*protocol.PurposeView, focus string) float64 {
	var v float64
	for _, r := range rows {
		var ps purposeSet
		ps.add(r)
		v += p.purposeValue(st, ps, 0, nil, true)
		if r.DamageToCreature > 0 {
			v += st.bestCreatureKill(r.DamageToCreature)
		}
		if r.DamageEachOpponent > 0 {
			v += float64(r.DamageEachOpponent*len(st.opps)) * p.cfg.DamageToOpponent
		}
		if r.LifeGain > 0 && st.myEval != nil {
			v += float64(r.LifeGain) * st.w.MarginalLife(st.myEval.Life)
		}
		if r.ExtraCombat > 0 {
			v += float64(r.ExtraCombat) * p.secondCombat(st, atk, focus)
		}
	}
	return v
}

// bestCreatureKill is the value of the best creature an opponent
// controls that `dmg` damage destroys. The catalog's target clause may be
// narrower (Glorybringer's non-Dragon); this is an estimate over the
// moves the enumerator already made legal.
func (st *state) bestCreatureKill(dmg int) float64 {
	best := 0.0
	for i := range st.view.Battlefield.Cards {
		c := &st.view.Battlefield.Cards[i]
		if !st.isOpponent(c.Controller) || !isCreature(c) || hasKeyword(c, "indestructible") {
			continue
		}
		if effectiveToughness(c) <= dmg {
			best = max(best, st.w.CombatValue(c))
		}
	}
	return best
}

// secondCombat is what an additional combat phase is worth: the attack
// value, against the focus seat, of each of the bot's OTHER untapped
// creatures that could attack in it, counted where it is positive.
func (p *Policy) secondCombat(st *state, atk *protocol.CardView, focus string) float64 {
	def := st.evals[focus]
	if def == nil {
		return 0
	}
	var v float64
	for i := range st.view.Battlefield.Cards {
		c := &st.view.Battlefield.Cards[i]
		if c.Controller != st.me || c.InstanceID == atk.InstanceID || !isCreature(c) ||
			c.Tapped || c.SummoningSick || hasKeyword(c, "defender") {
			continue
		}
		if a, _ := p.attackValue(st, c, def, 0, 0); a > 0 {
			v += a
		}
	}
	return v
}

// exertFree reports whether exerting atk costs the bot nothing (§9,
// §10): it has vigilance, or it won't untap during the bot's next untap
// step anyway (a marker naming the bot, a static restriction, a stun
// counter). A lethal push is the caller's third case.
func (st *state) exertFree(atk *protocol.CardView) bool {
	if hasKeyword(atk, "vigilance") {
		return true
	}
	if atk.Counters["stun"] > 0 {
		return true
	}
	if n := atk.NoUntap; n != nil && (n.Static || slices.Contains(n.Next, st.me)) {
		return true
	}
	return false
}

// exertCost is what an exert gives up when it is not free: the
// creature's best attack next turn, and its best block in each
// opponent's turn, both read off the board as it is now.
func (p *Policy) exertCost(st *state, atk *protocol.CardView) float64 {
	next := 0.0
	for _, o := range st.opps {
		if a, _ := p.attackValue(st, atk, o, 0, 0); a > next {
			next = a
		}
	}
	block := 0.0
	for _, o := range st.opps {
		best := 0.0
		for i := range st.view.Battlefield.Cards {
			c := &st.view.Battlefield.Cards[i]
			if c.Controller != o.ID || !isCreature(c) || c.Power <= 0 || hasKeyword(c, "defender") {
				continue
			}
			if !couldBlock(st, st.me, c, atk) {
				continue
			}
			v := 0.0
			if st.myEval != nil {
				v += float64(c.Power) * st.w.MarginalLife(st.myEval.Life)
			}
			if kills(atk, c) {
				v += st.w.CombatValue(c)
			}
			if kills(c, atk) {
				v -= st.w.CombatValue(atk)
			}
			best = max(best, v)
		}
		block += best
	}
	return p.cfg.ExertCostWeight * (next + block)
}

// exertTwinValue prices the exert twin of an attack against `def`: the
// attack with the pumped attacker, plus the gain, less the cost unless
// the exert is free or `push` (a lethal push on that seat) makes it so.
func (p *Policy) exertTwinValue(st *state, atk *protocol.CardView, def *SeatEval, declared, declaredPower int, push bool, focus string) (float64, string) {
	rows := st.exertRows(atk)
	v, reason := p.attackValue(st, pumped(atk, rows), def, declared, declaredPower)
	gain := p.exertGain(st, atk, rows, focus)
	cost := 0.0
	if !push && !st.exertFree(atk) {
		cost = p.exertCost(st, atk)
	}
	return v + gain - cost, fmt.Sprintf("%s, exerted (gain %.2f, cost %.2f)", reason, gain, cost)
}

// exertTwin finds the exert twin of attack move `plain` (same attacker,
// same target) in moves, or -1.
func exertTwin(moves []legal.Move, plain attackParams) int {
	for i := range moves {
		if !exertMove(moves[i]) {
			continue
		}
		ap := decode[attackParams](moves[i].Params)
		if ap.Attacker == plain.Attacker && ap.Target == plain.Target {
			return i
		}
	}
	return -1
}

// isOpponent reports whether seat is one of the live opponents.
func (st *state) isOpponent(seat string) bool {
	for _, o := range st.opps {
		if o.ID == seat {
			return true
		}
	}
	return false
}

// pumps reports whether any of rows pumps or shields its source, which
// only helps the attack it is declared with.
func pumps(rows []*protocol.PurposeView) bool {
	for _, r := range rows {
		if r.Pump != nil || r.PreventCombatDamageToSelf {
			return true
		}
	}
	return false
}
