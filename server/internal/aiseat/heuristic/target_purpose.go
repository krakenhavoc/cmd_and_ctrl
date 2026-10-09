package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// target_purpose.go prices a target by what the spell or ability does
// TO it, when the catalog declares that (ADR 0126's amendment of
// 2026-10-08, owner answers 1 to 4 and 6).
//
// targetsValue's sign convention is "a target is hit": a player who is
// not the bot is priced as an attack, the bot itself as a penalty, and
// every opposing creature as removed. That is wrong for a spell that
// gives its target something. Prismari Command's "Target player draws
// two cards, then discards two cards" at an opponent was priced as
// burn, and at the bot as a mistake. And it is wrong for damage that
// kills nothing: 2 damage at a 2/4 was priced as removing it.
//
// A declared target entry (PurposeView.targets, keyed by the clause's
// slot) says what happens to the pick.
//
// THE GIFTS (PriceTargetPurposes, A1 and B1): cards drawn and
// discarded, tokens made, life gained and lost. Those amounts are the
// seat's strength change, priced the way purposeValue prices the same
// amounts for the bot:
//
//	x = Hand × draws − DiscardWeight × discards + TokenWeight × tokens
//	    + the strength change of the life gained and lost
//	    (+ the bot's own discard payoffs, for the bot)
//
// and the move is worth the change in ScoreEval when each target seat's
// strength moves by its x. For the bot that is +x. For an opponent it
// is −x weighed by OpponentMean and OpponentMax, which is how §4 prices
// a sweep: −1.5x at a two-seat table, less for one of three opponents.
//
// THE DAMAGE (DamageByLethality, C1 and D1):
//
//   - at a player, DamageToOpponent per point is that seat's strength
//     change, through the same weights (the bot's own life at
//     MarginalLife), and LethalBonus when the life the move takes from
//     the seat reaches its life total (CR 704.5a);
//   - at a creature, removal if the damage kills it (damageKills: CR
//     120.6, 702.2b, 702.12b, 702.16e), DamageChip of removal if it
//     survives, because marked damage is removed in cleanup (CR 514.2);
//     the bot's own creature costs its value if it dies and nothing if
//     it survives;
//   - at a planeswalker, the share of its loyalty removed (CR 120.3c),
//     all of it at or above its loyalty (CR 704.5i);
//   - at a battle, or a pick the view does not show, today's price.
//
// A pick with no entry keeps targetsValue's price exactly as before.
// Nothing is inferred.

// giftAmounts is what the priced entries give one seat.
type giftAmounts struct {
	draws, discards, tokens, lifeGain, lifeLoss, damage int
}

// givesPlayer reports whether an entry names an amount only a player
// can be given (game.TargetPurpose.HasPlayerAmount on the wire).
func givesPlayer(e *protocol.TargetPurposeView) bool {
	return e != nil && (e.Draws != 0 || e.Discards != 0 || e.Tokens != 0 || e.LifeGain != 0 || e.LifeLoss != 0)
}

// entryAt is the entry for clause `slot` in a statement's purpose, nil
// when it declares none.
func entryAt(p *protocol.PurposeView, slot int) *protocol.TargetPurposeView {
	if p == nil || p.Targets == nil {
		return nil
	}
	for i := range *p.Targets {
		if e := &(*p.Targets)[i]; e.Slot == slot {
			return e
		}
	}
	return nil
}

// castEntryFor is the declared entry for one pick of a cast: the chosen
// mode's statement for a modal cast (t.Mode is the occurrence in
// cp.Modes), else the claimed alternative cost's when it declares
// entries, else the card's own.
func castEntryFor(c *protocol.CardView, cp castParams, t targetRef) *protocol.TargetPurposeView {
	if c == nil {
		return nil
	}
	if c.Modes != nil && len(cp.Modes) > 0 {
		if t.Mode < 0 || t.Mode >= len(cp.Modes) {
			return nil
		}
		opt := cp.Modes[t.Mode]
		if opt < 0 || opt >= len(c.Modes.Options) {
			return nil
		}
		return entryAt(c.Modes.Options[opt].Purpose, t.Slot)
	}
	if t.Mode != 0 {
		return nil
	}
	if cp.AlternativeCost != "" {
		for i := range c.AlternativeCosts {
			if ac := &c.AlternativeCosts[i]; ac.Key == cp.AlternativeCost && ac.Purpose != nil && ac.Purpose.Targets != nil {
				return entryAt(ac.Purpose, t.Slot)
			}
		}
	}
	return entryAt(c.Purpose, t.Slot)
}

// rowEntryFor is the declared entry for one pick of an activated row.
func rowEntryFor(src *protocol.CardView, index int, t targetRef) *protocol.TargetPurposeView {
	if t.Mode != 0 {
		return nil
	}
	return entryAt(rowPurpose(src, index), t.Slot)
}

// pricedTargetsValue prices a move's targets: each pick whose declared
// entry this Config prices goes through giftsValue (a player) or
// damageCardValue (a permanent), and every other pick through
// targetsValue as before. With DamageByLethality off, an entry's damage
// keeps targetsValue's price for its pick on top of its gifts. `self`
// is the card being cast, left out of the bot's discard candidates, and
// `source` the damage's source, whose colours and deathtouch the kill
// test reads. The second result reports whether any pick was priced by
// its entry, which makes the move purpose-priced (purposeSet's
// targetsPriced). With PriceTargetPurposes off it is targetsValue.
// Either way, what a removal's declared `returns` hand the target's
// controller is taken off (NetRemoval, net_removal.go); a return alone
// does not make the move purpose-priced.
func (p *Policy) pricedTargetsValue(st *state, targets []targetRef, entryOf func(targetRef) *protocol.TargetPurposeView, self, source *protocol.CardView) (float64, bool) {
	v, priced := p.pricedTargetsGross(st, targets, entryOf, self, source)
	return v - p.removalReturnsValue(st, targets, entryOf), priced
}

// pricedTargetsGross is pricedTargetsValue before the returns.
func (p *Policy) pricedTargetsGross(st *state, targets []targetRef, entryOf func(targetRef) *protocol.TargetPurposeView, self, source *protocol.CardView) (float64, bool) {
	if !p.cfg.PriceTargetPurposes {
		return st.targetsValue(p.cfg, targets), false
	}
	lethality := p.cfg.DamageByLethality
	var rest []targetRef
	var gifts map[string]*giftAmounts
	var cards float64
	priced := false
	for _, t := range targets {
		e := entryOf(t)
		if t.Kind != "player" {
			if lethality && e != nil && e.Damage > 0 {
				if v, ok := p.damageCardValue(st, t.ID, e.Damage, source); ok {
					cards += v
					priced = true
					continue
				}
			}
			rest = append(rest, t)
			continue
		}
		damage := lethality && e != nil && e.Damage > 0
		if !givesPlayer(e) && !damage {
			rest = append(rest, t)
			continue
		}
		if gifts == nil {
			gifts = map[string]*giftAmounts{}
		}
		g := gifts[t.ID]
		if g == nil {
			g = &giftAmounts{}
			gifts[t.ID] = g
		}
		g.draws += e.Draws
		g.discards += e.Discards
		g.tokens += e.Tokens
		g.lifeGain += e.LifeGain
		g.lifeLoss += e.LifeLoss
		if damage {
			g.damage += e.Damage
		} else if e.Damage > 0 {
			rest = append(rest, t)
		}
		priced = true
	}
	if !priced {
		return st.targetsValue(p.cfg, targets), false
	}
	return st.targetsValue(p.cfg, rest) + cards + p.giftsValue(st, gifts, self), true
}

// giftsValue is the change in ScoreEval when each seat in gifts gets
// what it is given (owner answers 2 and 4): +x for the bot, the
// opposition weights' share of −x for an opponent, and LethalBonus for
// a seat the move puts at 0 life or less.
func (p *Policy) giftsValue(st *state, gifts map[string]*giftAmounts, self *protocol.CardView) float64 {
	if len(gifts) == 0 || st.evals[st.me] == nil {
		return 0
	}
	after := make(map[string]*SeatEval, len(st.evals))
	for id, e := range st.evals {
		after[id] = e
	}
	var lethal float64
	for seat, g := range gifts {
		e := st.evals[seat]
		if e == nil || e.Eliminated {
			continue
		}
		x := st.w.Hand*float64(g.draws) - p.cfg.DiscardWeight*float64(g.discards) +
			p.cfg.TokenWeight*float64(g.tokens) + st.w.lifeChange(e, g.lifeGain-g.lifeLoss)
		if seat == st.me {
			// Mary Read's Treasure for the Island the loot pitches, as
			// purposeValue counts it for an untargeted loot.
			x += st.resolutionDiscardPayoff(p.cfg, g.discards, self)
			x -= st.w.MarginalLife(e.Life) * float64(g.damage)
		} else {
			x -= p.cfg.DamageToOpponent * float64(g.damage)
		}
		// CR 704.5a: a player at 0 or less life loses. Only a move that
		// deals damage is read this way (DamageByLethality); a life loss
		// alone keeps the low-life penalty lifeChange prices.
		if g.damage > 0 && !e.CantLoseLife && g.damage+g.lifeLoss-g.lifeGain >= e.Life {
			if seat == st.me {
				lethal -= p.cfg.LethalBonus
			} else {
				lethal += p.cfg.LethalBonus
			}
		}
		cp := *e
		cp.Strength += x
		after[seat] = &cp
	}
	return st.w.ScoreEval(after, st.me) - st.w.ScoreEval(st.evals, st.me) + lethal
}

// damageCardValue is a declared damage entry's price at permanent `id`
// (C1, owner answers 3 and 6). False for a pick it does not price: one
// the view does not show on the battlefield, a battle, or a permanent
// damage is not dealt to; targetsValue prices those as before.
func (p *Policy) damageCardValue(st *state, id string, dmg int, source *protocol.CardView) (float64, bool) {
	c := st.bf[id]
	if c == nil {
		return 0, false
	}
	mine := c.Controller == st.me
	switch {
	case isCreature(c):
		value := st.w.CreatureValue(c)
		if damageKills(source, c, dmg) {
			if mine {
				return -value, true
			}
			// #2679: a commander comes back from the command zone.
			value = st.commanderRemovalValue(p.cfg, c, value)
			return value * p.cfg.RemovalConfidence * st.leaderBoost(p.cfg, c.Controller), true
		}
		if mine {
			return 0, true
		}
		return p.cfg.DamageChip * value * p.cfg.RemovalConfidence * st.leaderBoost(p.cfg, c.Controller), true
	case isType(c, "planeswalker"):
		share := 0.0
		if !protectedFrom(c, source) {
			if loyalty := c.Counters["loyalty"]; loyalty <= dmg {
				share = 1
			} else {
				share = float64(dmg) / float64(loyalty)
			}
		}
		value := st.permanentValue(c)
		if mine {
			return -share * value, true
		}
		if share >= 1 {
			// #2679: a commander comes back from the command zone.
			value = st.commanderRemovalValue(p.cfg, c, value)
		}
		return share * value * p.cfg.RemovalConfidence * st.leaderBoost(p.cfg, c.Controller), true
	}
	return 0, false
}

// lifeChange is the change in a seat's Strength when its life moves by
// d: the linear term and the low-life penalty, clamped at 0 life for a
// seat that can't lose to it, exactly as Evaluate computes Strength.
func (w Weights) lifeChange(e *SeatEval, d int) float64 {
	if d == 0 {
		return 0
	}
	at := func(life int) float64 {
		danger := life
		if e.CantLoseLife && danger < 0 {
			danger = 0
		}
		return w.Life*float64(life) - w.lifeDanger(danger)
	}
	return at(e.Life+d) - at(e.Life)
}
