package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// target_purpose.go prices a target by what the spell or ability does
// TO it, when the catalog declares that (ADR 0126's amendment of
// 2026-10-08, A1 and B1, owner answers 1 and 2).
//
// targetsValue's sign convention is "a target is hit": a player who is
// not the bot is priced as an attack, the bot itself as a penalty. That
// is wrong for a spell that gives its target something. Prismari
// Command's "Target player draws two cards, then discards two cards" at
// an opponent was priced as burn, and at the bot as a mistake.
//
// A declared target entry (PurposeView.targets, keyed by the clause's
// slot) says what happens to the pick: cards drawn and discarded,
// tokens made, life gained and lost. Those amounts are the seat's
// strength change, priced the way purposeValue prices the same amounts
// for the bot:
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
// Damage is not priced here. An entry's `damage` keeps targetsValue's
// price for its pick until DamageByLethality (the amendment's PR 4),
// and a pick with no entry, or an entry with no player amount, keeps
// targetsValue's price exactly as before. Nothing is inferred.

// giftAmounts is what the priced entries give one seat.
type giftAmounts struct {
	draws, discards, tokens, lifeGain, lifeLoss int
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
// entry gives a player something is priced by giftsValue, and every
// other pick by targetsValue as before. An entry that also deals damage
// keeps targetsValue's price for its pick on top (the damage half,
// PR 4's). `self` is the card being cast, left out of the bot's discard
// candidates. The second result reports whether any pick was priced by
// its entry, which makes the move purpose-priced (purposeSet's
// targetsPriced). With PriceTargetPurposes off it is targetsValue.
func (p *Policy) pricedTargetsValue(st *state, targets []targetRef, entryOf func(targetRef) *protocol.TargetPurposeView, self *protocol.CardView) (float64, bool) {
	if !p.cfg.PriceTargetPurposes {
		return st.targetsValue(p.cfg, targets), false
	}
	var rest []targetRef
	var gifts map[string]*giftAmounts
	for _, t := range targets {
		e := entryOf(t)
		if t.Kind != "player" || !givesPlayer(e) {
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
		if e.Damage > 0 {
			rest = append(rest, t)
		}
	}
	if gifts == nil {
		return st.targetsValue(p.cfg, targets), false
	}
	return st.targetsValue(p.cfg, rest) + p.giftsValue(st, gifts, self), true
}

// giftsValue is the change in ScoreEval when each seat in gifts gets
// what it is given (owner answer 2): +x for the bot, the opposition
// weights' share of −x for an opponent.
func (p *Policy) giftsValue(st *state, gifts map[string]*giftAmounts, self *protocol.CardView) float64 {
	if st.evals[st.me] == nil {
		return 0
	}
	after := make(map[string]*SeatEval, len(st.evals))
	for id, e := range st.evals {
		after[id] = e
	}
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
		}
		cp := *e
		cp.Strength += x
		after[seat] = &cp
	}
	return st.w.ScoreEval(after, st.me) - st.w.ScoreEval(st.evals, st.me)
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
