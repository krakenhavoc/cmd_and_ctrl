package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// cant_block.go — #4977: "can't block this turn" on an opponent's
// creature, priced by the attack it opens before the defender's blocks.
//
// A harm until end of turn that prints no amount (Tragic Slip's -X/-X)
// keeps the old reading, an opposing pick removed. "Can't block" was the
// same bare harm, so Daring Discovery's "up to three target creatures
// can't block this turn" was priced as three removal spells, and
// Untimely Malfunction's third mode as two. The catalog now declares the
// restriction (Pump.CantBlock, `cant_block` on the wire), and with
// Config.PriceCantBlock such a pick of an opponent's creature is worth,
// together with the move's other such picks, what the bot's attack gains
// when those creatures cannot block:
//
//   - only on the bot's own turn, before blockers are declared: in its
//     first main phase and beginning of combat for every creature that
//     could still attack, and in its declare attackers step for the ones
//     already attacking. Once blocks are declared a "can't block" changes
//     nothing (CR 509.1b), and on another seat's turn the bot is not the
//     one attacking;
//   - per opponent, the swing the bot would make at that seat (swingAt:
//     the attackers already declared, then each one still able to join,
//     biggest first, while attackValue says it is worth sending) with the
//     picks unable to block, less the same swing with them able, plus
//     LethalBonus when the picks are what turns the all-in swing into a
//     kill (lethalPushAs). The best opponent's gain, never below zero.
//
// A pick whose declared -N/-N kills it is still removal, and a pick of the
// bot's own creature is still the harm on its own side #4302 prices.
//
// The switch ships OFF. In #4977's arena (cantblock-battle, two seats with
// the price against two without, seeds 2000-2047) the seats with it won
// 19 of 96 seat-games against 29 of 96: they held Daring Discovery for an
// attack and then mostly cast it at one creature, where the removal price
// casts it at once at the two or three biggest. `heuristic-cantblock`
// seats it on in the arena.

// cantBlockPick reports whether pick t, declared by entry e, is one this
// Config prices by the attack it opens: a can't-block harm on an
// opponent's creature that its declared shrink does not kill.
func (p *Policy) cantBlockPick(st *state, t targetRef, e *protocol.TargetPurposeView) (*protocol.CardView, bool) {
	if !p.cfg.PriceCantBlock || !p.cfg.PriceTemporaryEffects || e == nil || t.Kind == "player" {
		return nil, false
	}
	if e.UntilEndOfTurn != targetHarm || e.Pump == nil || !e.Pump.CantBlock {
		return nil, false
	}
	c := st.bf[t.ID]
	if c == nil || c.Controller == st.me || shrinkKills(c, e.Pump) {
		return nil, false
	}
	return c, true
}

// declaresCantBlock reports whether entry e is a can't-block harm this
// Config prices. A cast of Daring Discovery that names none of its "up to
// three" creatures is then priced by what it declares, its discover, and
// not by the mana-value proxy, which would outbid the same cast at the
// creatures that block.
func (p *Policy) declaresCantBlock(e *protocol.TargetPurposeView) bool {
	return p.cfg.PriceCantBlock && p.cfg.PriceTemporaryEffects && p.cfg.PriceTargetPurposes &&
		e != nil && e.UntilEndOfTurn == targetHarm && e.Pump != nil && e.Pump.CantBlock
}

// cantBlockValue is the attack the picks open, all of them at once: what
// a move's can't-block picks are worth together.
func (p *Policy) cantBlockValue(st *state, picks []*protocol.CardView) float64 {
	if len(picks) == 0 || !st.myTurn {
		return 0
	}
	switch st.step {
	case "precombat_main", "begin_combat", "declare_attackers":
	default:
		return 0
	}
	best := 0.0
	for _, o := range st.opps {
		mine := false
		for _, c := range picks {
			if c.Controller == o.ID && isCreature(c) && !c.Tapped {
				mine = true
			}
		}
		if !mine {
			continue
		}
		before := p.swingAt(st, o)
		killsBefore := p.lethalPushAs(st, o, nil)
		st.noBlock = map[string]bool{}
		for _, c := range picks {
			st.noBlock[c.InstanceID] = true
		}
		after := p.swingAt(st, o)
		killsAfter := p.lethalPushAs(st, o, nil)
		st.noBlock = nil
		gain := after - before
		if killsAfter && !killsBefore {
			gain += p.cfg.LethalBonus
		}
		best = max(best, gain)
	}
	return best
}

// swingAt is the bot's attack at seat def as its attack planner would
// make it, one creature at a time: the creatures already attacking def,
// then, before attackers are declared, each one still able to attack,
// biggest first, sent while attackValue says it is worth it. The sum of
// their attack values.
func (p *Policy) swingAt(st *state, def *SeatEval) float64 {
	joining := st.step == "precombat_main" || st.step == "begin_combat"
	var committed, joinable []*protocol.CardView
	for i := range st.view.Battlefield.Cards {
		c := &st.view.Battlefield.Cards[i]
		if c.Controller != st.me || !isCreature(c) {
			continue
		}
		switch {
		case c.AttackingTarget == def.ID:
			committed = append(committed, c)
		case joining && c.AttackingTarget == "" && !c.Tapped && !c.SummoningSick &&
			!hasKeyword(c, "defender") && !hasRestriction(c, "cant_attack"):
			joinable = append(joinable, c)
		}
	}
	total := 0.0
	declared, power := 0, 0
	send := func(c *protocol.CardView, v float64) {
		total += v
		if p.cfg.FocusNeedsValue && combatPower(c) <= 0 {
			return
		}
		declared++
		power += combatPower(c)
	}
	for _, c := range committed {
		v, _ := p.attackValue(st, c, def, declared, power)
		send(c, v)
	}
	for _, i := range byPower(joinable) {
		c := joinable[i]
		if v, _ := p.attackValue(st, c, def, declared, power); v > 0 {
			send(c, v)
		}
	}
	return total
}
