package heuristic

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// rocknow.go is ADR 0136's amendment of 2026-10-09, "a rock against a
// spell over two turns" (Config.PlanRockTwoTurns), and ADR 0126 §2's
// amendment of the same day, "an idle late rock" (Config.IdleLateRocks).
//
// The two-turn comparison. The turn plan (plan.go) prices one turn. When
// this turn's mana pays for a mana source or for the turn's chosen line
// but not both, the plan takes the line, because a two-drop prices above
// a Signet. A player asks what each order buys over two turns instead:
//
//	A = the best set holding the source, this turn
//	    + PlanNextTurnDiscount × the best set next turn's mana buys with
//	      the source on the battlefield
//	B = the line chosen this turn
//	    + PlanNextTurnDiscount × the best set next turn's mana buys
//	      without it
//
// and takes A when it is worth more. Only the bot's hand, its command
// zone and the mana model (§2) are read: no hidden information and no
// opponent is simulated. Next turn's mana (nextTurnMana) is every mana
// permanent the bot controls, untapped in its untap step, plus one land
// if a land is in hand, plus what this turn's casts leave on the
// battlefield. Each turn's set is valued as §3 values a plan, the ramp
// premium shared within it, and next turn's sets may hold any card in
// hand or in the command zone, whether or not it is castable now.

// twoTurnLine is one order of the comparison: what is cast this turn,
// and what next turn then buys.
type twoTurnLine struct {
	now     []*planCandidate
	nowVal  float64
	next    []*planCandidate
	nextVal float64
	// held is the members of now that §5 holds for the end step, as bits
	// in cands (Config.PlanHoldInstants).
	held  uint
	cands []*planCandidate
}

func (l twoTurnLine) total(discount float64) float64 { return l.nowVal + discount*l.nextVal }

// decideRockNow is the two-turn comparison. line is the turn's chosen
// line, in move indices (the plan's members, or the one move
// decideGeneral would take), worth lineVal; pl is the plan search's
// result, which carries this window's candidates. It returns the first
// move of the order holding the mana source when that order is worth
// more over the two turns and clears the window's bar.
func (p *Policy) decideRockNow(ctx context.Context, st *state, moves []legal.Move, vals []float64,
	pl turnPlan, line []int, lineVal, threshold float64, leftover bool) (aiseat.Decision, []aiseat.PlanMember, bool) {
	if !p.cfg.PlanTurnMana || !p.cfg.PlanRockTwoTurns || !st.sorcerySpeed || len(line) == 0 || ctx.Err() != nil {
		return aiseat.Decision{}, nil, false
	}
	cands := pl.cands
	if len(cands) < 2 {
		return aiseat.Decision{}, nil, false
	}
	// The line's members, as bits in cands. A line holding a move the
	// plan cannot model (an activation, a cast with a sacrifice other
	// than a land swap that nets lands, planEligibleIn) is left alone.
	var lineMask uint
	for _, idx := range line {
		k := p.candidateFor(st, cands, moves, idx)
		if k < 0 {
			return aiseat.Decision{}, nil, false
		}
		lineMask |= 1 << k
	}
	base := st.manaAvailable(p.cfg.PlanFilterLands)
	want := st.rampWants()
	sources := st.manaSourcesTotal()
	units := make([]uint8, 0, 2*len(base)+8)

	var lineB *twoTurnLine
	var best *twoTurnLine
	for r, rock := range cands {
		bit := uint(1) << r
		// A mana source whose price carries ADR 0126 §2's premium: the
		// deficit is open. It is not in the line, and the turn's mana
		// pays for it or for the line, not both.
		if rock.premium <= 0 || lineMask&bit != 0 {
			continue
		}
		if setFeasible(base, cands, lineMask|bit, units) {
			continue
		}
		mask, val, ord, held, ok := p.bestSet(ctx, setSearch{
			cands: cands, base: base, want: want, sources: sources,
			must: bit, minSize: 1, trustSingles: true, hold: p.cfg.PlanHoldInstants,
		})
		if !ok {
			continue
		}
		a := &twoTurnLine{now: membersOf(cands, ord, mask, held), nowVal: val, held: held, cands: cands}
		if a.nowVal <= p.setBar(st, moves, a.now, threshold, leftover) {
			continue
		}
		a.next, a.nextVal = p.nextTurnBest(ctx, st, moves, vals, a.now)
		if lineB == nil {
			b := &twoTurnLine{nowVal: lineVal}
			for _, idx := range line {
				b.now = append(b.now, cands[p.candidateFor(st, cands, moves, idx)])
			}
			b.next, b.nextVal = p.nextTurnBest(ctx, st, moves, vals, b.now)
			lineB = b
		}
		if a.total(p.cfg.PlanNextTurnDiscount) <= lineB.total(p.cfg.PlanNextTurnDiscount)+planEpsilon {
			continue
		}
		if best == nil || a.total(p.cfg.PlanNextTurnDiscount) > best.total(p.cfg.PlanNextTurnDiscount)+planEpsilon {
			best = a
		}
	}
	if best == nil {
		return aiseat.Decision{}, nil, false
	}
	a2 := turnPlan{members: best.now, cands: best.cands, held: best.held}
	d := aiseat.Decision{Index: best.now[0].index, Reason: twoTurnReason(*best, *lineB, p.cfg.PlanNextTurnDiscount)}
	if a2.isHeld(best.now[0]) {
		// §5: held members come last, so every member of the order is
		// held for the end step, and nothing is cast now.
		d.Index = passOrDecline(moves)
	}
	return d, planMembers(moves, a2), true
}

// candidateFor is the index in cands of the card move idx casts, -1
// when the move is not a plan candidate.
func (p *Policy) candidateFor(st *state, cands []*planCandidate, moves []legal.Move, idx int) int {
	if idx < 0 || idx >= len(moves) || !p.planEligibleIn(st, moves[idx]) {
		return -1
	}
	id := decode[castParams](moves[idx].Params).InstanceID
	for k, c := range cands {
		if c.card.InstanceID == id {
			return k
		}
	}
	return -1
}

// twoTurnReason is the decision's reason: "two turns: Arcane Signet now,
// then Avenger of Zendikar (+2.75) over Undead Butler now, then nothing
// (+1.59)".
func twoTurnReason(a, b twoTurnLine, discount float64) string {
	return fmt.Sprintf("two turns: %s over %s", a.describe(discount), b.describe(discount))
}

func (l twoTurnLine) describe(discount float64) string {
	return fmt.Sprintf("%s now, then %s (+%.2f)", candidateNames(l.now), candidateNames(l.next), l.total(discount))
}

func candidateNames(cs []*planCandidate) string {
	if len(cs) == 0 {
		return "nothing"
	}
	names := make([]string, 0, len(cs))
	for _, c := range cs {
		names = append(names, c.card.Name)
	}
	return strings.Join(names, " → ")
}

// nextTurnBest is the best set next turn's mana buys after `now` is cast
// this turn, and its value (never below nothing: the bot may pass). The
// cards are the bot's hand and command zone less `now`, each priced as
// it is now: by its best plan-eligible cast on offer, or, for a card the
// bot cannot cast this turn, by a cast from its zone that names no
// target. A card offered only with a cost the plan cannot hold (§1) is
// left out.
func (p *Policy) nextTurnBest(ctx context.Context, st *state, moves []legal.Move, vals []float64, now []*planCandidate) ([]*planCandidate, float64) {
	gone := map[string]bool{}
	for _, c := range now {
		gone[c.card.InstanceID] = true
	}
	var pool []*planCandidate
	offered := map[string]bool{}
	for i := range moves {
		if moves[i].Kind == legal.KindCast {
			offered[decode[castParams](moves[i].Params).InstanceID] = true
		}
	}
	for _, c := range p.eligibleCasts(st, moves, vals) {
		if !gone[c.card.InstanceID] {
			pool = append(pool, c)
		}
	}
	if st.seat != nil {
		add := func(c *protocol.CardView, zone string) {
			if isLand(c) || gone[c.InstanceID] || offered[c.InstanceID] || c.ManaCost == "" {
				return
			}
			raw, err := json.Marshal(castParams{InstanceID: c.InstanceID, FromZone: zone})
			if err != nil {
				return
			}
			m := legal.Move{Kind: legal.KindCast, Type: legal.TypeCastSpell, Params: raw, Label: "Cast " + c.Name}
			cost, ok := st.planCastCost(m, c, castParams{FromZone: zone})
			if !ok {
				return
			}
			v, _ := p.valueOf(st, m)
			pool = append(pool, &planCandidate{index: -1, move: m, card: c, value: v, cost: cost})
		}
		for i := range st.seat.Hand.Cards {
			add(&st.seat.Hand.Cards[i], "hand")
		}
		for i := range st.seat.Command.Cards {
			add(&st.seat.Command.Cards[i], "command")
		}
	}
	if len(pool) == 0 {
		return nil, 0
	}
	// eligibleCasts made fresh candidates, so annotating them leaves
	// this turn's plan alone.
	cands := p.rankCandidates(st, pool)
	var want []rampWant
	for _, w := range st.rampWants() {
		if !gone[w.id] {
			want = append(want, w)
		}
	}
	sources := st.manaSourcesTotal()
	for _, c := range now {
		sources += c.amount
	}
	mask, val, ord, _, ok := p.bestSet(ctx, setSearch{
		cands: cands, base: st.nextTurnMana(p.cfg.PlanFilterLands, now),
		want: want, sources: sources, minSize: 1,
	})
	if !ok || val <= 0 {
		return nil, 0
	}
	return membersOf(cands, ord, mask, 0), val
}

// nextTurnMana is the mana the bot can make in its next turn's main
// phase (the amendment of 2026-10-09): every mana permanent it controls,
// since each untaps in its untap step and a creature has been under its
// control since the turn began (CR 302.6); one land if a land is in hand,
// for the next land drop; and the mana sources this turn's casts (`now`)
// leave on the battlefield, a ramp spell's lands included. The pool
// empties before then and is not counted.
func (st *state) nextTurnMana(filterLands bool, now []*planCandidate) []uint8 {
	var srcs []manaSource
	for i := range st.view.Battlefield.Cards {
		c := &st.view.Battlefield.Cards[i]
		if c.Controller != st.me {
			continue
		}
		if src, ok := manaSourceOf(c, filterLands); ok {
			srcs = append(srcs, src)
		}
	}
	if st.seat != nil {
		for i := range st.seat.Hand.Cards {
			c := &st.seat.Hand.Cards[i]
			if !isLand(c) {
				continue
			}
			if src, ok := manaSourceOf(c, filterLands); ok {
				srcs = append(srcs, src)
			} else if len(c.ManaAbilities) == 0 {
				srcs = append(srcs, manaSource{units: []uint8{manaAnyColor}})
			}
			break
		}
	}
	for _, c := range now {
		if isPermanentSpell(c.card) && !isLand(c.card) && len(c.card.ManaAbilities) > 0 {
			if src, ok := manaSourceOf(c.card, filterLands); ok {
				srcs = append(srcs, src)
			}
		}
		for k := 0; k < c.lands; k++ {
			srcs = append(srcs, manaSource{units: []uint8{manaAnyColor}})
		}
	}
	return unitsOf(srcs)
}

// unitsOf is the mana a list of sources makes once each is activated:
// the plain sources first, then the filters, fed as the engine would.
func unitsOf(srcs []manaSource) []uint8 {
	var units []uint8
	var filters []manaSource
	for _, s := range srcs {
		if s.input > 0 {
			filters = append(filters, s)
			continue
		}
		units = append(units, s.units...)
	}
	for _, f := range filters {
		units = addSource(units, f, nil)
	}
	return units
}

// idleRock is ADR 0126 §2's amendment of 2026-10-09: in the turn's last
// main-phase window, a mana source with no open deficit (its §2 price is
// below zero on purpose) is priced at LeftoverThreshold when its mana
// would otherwise go unused, that is, when no other move on offer is
// priced above LeftoverThreshold. A mana permanent is never worthless:
// past RampWantCap its mana still pays a commander's tax or an X. It
// returns the move, -1 for none. vals is valueOf for each move.
func (p *Policy) idleRock(st *state, moves []legal.Move, vals []float64) int {
	if !p.cfg.IdleLateRocks || !p.cfg.LeftoverWindows || !st.leftover || !st.sorcerySpeed {
		return -1
	}
	var rocks []int
	for i := range moves {
		switch moves[i].Kind {
		case legal.KindPass, legal.KindFinishBlocks:
			continue
		}
		if p.isIdleRock(st, moves[i]) {
			rocks = append(rocks, i)
			continue
		}
		if vals[i] > p.cfg.LeftoverThreshold {
			return -1
		}
	}
	if len(rocks) == 0 {
		return -1
	}
	// The source that makes the most mana, then the dearer price.
	sort.SliceStable(rocks, func(i, j int) bool {
		a := st.castSource(decode[castParams](moves[rocks[i]].Params).InstanceID)
		b := st.castSource(decode[castParams](moves[rocks[j]].Params).InstanceID)
		if ra, rb := repeatableMana(a), repeatableMana(b); ra != rb {
			return ra > rb
		}
		return vals[rocks[i]] > vals[rocks[j]]
	})
	return rocks[0]
}

// isIdleRock reports whether m casts a nonland permanent that makes
// repeatable mana, with no ramp premium (no open deficit), paying only
// mana (ADR 0126 §5's costsOnlyManaAndTaps).
func (p *Policy) isIdleRock(st *state, m legal.Move) bool {
	if m.Kind != legal.KindCast {
		return false
	}
	c := st.castSource(decode[castParams](m.Params).InstanceID)
	if c == nil || isLand(c) || !isPermanentSpell(c) || repeatableMana(c) <= 0 {
		return false
	}
	return p.rampPremium(st, c) <= 0 && p.leftoverEligible(st, m)
}
