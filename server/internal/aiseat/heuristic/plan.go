package heuristic

import (
	"context"
	"fmt"
	"math/bits"
	"sort"
	"strconv"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// plan.go is ADR 0136's turn plan (PR 4): in its own main phase with an
// empty stack, the bot asks what the rest of this turn's mana buys, not
// only which one move is worth most.
//
//   - The members (§1, owner answer 3) are cast moves whose only costs
//     are mana and the card, and with Config.PlanHoldInstants a land
//     swap that nets lands (Harrow, planEligibleIn). Each card
//     contributes its best such move.
//     The PlanMaxCards cards whose best move prices highest are the
//     candidates, so the search sees at most 2^PlanMaxCards sets.
//   - The mana model (§2, owner answers 1, 2 and 4) is the bot's own,
//     from the view: the floating pool and each untapped permanent it
//     controls with a repeatable mana ability it can activate now. A
//     cast costs the total mana its move charges (legal.MoveCost.Mana).
//     A rock, a hasty dork, or a ramp spell whose lands enter untapped
//     (PurposeView.lands_untapped) adds its mana for the members after
//     it.
//   - A set's value (§3) is the sum of its members' prices with the ramp
//     premium shared: each member's own premium comes out, and one
//     premium for the set goes back, measured against the deficit of
//     the cards NOT in the set.
//   - The order (§4, owner answer 5): mana first, then draws and tutors,
//     cheapest first, then the rest, highest value first. A permanent
//     that grants extra land drops orders as mana (the amendment of
//     2026-10-08, Config.PlanLandDropsAsRamp). The bot makes the first
//     move; the next window plans again. Nothing is stored.
//   - The commander tax (§6, owner answer 7) is in the move's mana, so a
//     taxed commander competes with whatever else that mana buys.
//
// The model is advisory. A set of one is never searched: the engine
// offered the move, and its answer outranks the model's. If the model
// is wrong about a second member, that member is not offered in the
// next window and the bot re-plans. The arena counts those plan misses
// (botarena/turnmana.go).
//
// Instants (§5, owner answer 6, PR 5): with Config.PlanHoldInstants an
// instant-speed member is HELD for the end step before the bot's turn
// unless a later member needs its mana or it draws and leaves mana
// after it. A held member keeps its value and its mana in the plan, and
// is paid after every member cast this turn; it is not cast now. When
// every member left is held the bot passes, and ADR 0126 §5 casts it in
// the end step before the bot's turn (holdinstants.go).

// manaUnit is one mana in the model. Its low byte is the symbols it can
// pay for any cost: one bit per colour and {C}, and manaPain. A unit
// from a source with a restricted row (manaClasses) also carries, above
// that, the colours the restricted row makes and the row's restriction
// class: those colours pay only for a cost that admits the class
// (manaCost.admits).
type manaUnit uint32

// Colour masks for one mana in the model.
const (
	manaW manaUnit = 1 << iota
	manaU
	manaB
	manaR
	manaG
	manaC

	// manaPain marks a mana whose source costs life or deals damage to
	// its controller (Mana Confluence, Ancient Tomb, City of Brass). It
	// is not a colour: the engine's auto-tapper spends such a source
	// only once nothing painless is left, and the model does the same.
	manaPain manaUnit = 1 << 7

	manaAnyColor = manaW | manaU | manaB | manaR | manaG
	// manaAny pays any symbol at all, {C} included (a snow symbol's
	// stand-in).
	manaAny = manaAnyColor | manaC

	// manaRestrictedShift is where a unit keeps the symbols its
	// restricted row pays (manaAny's layout, shifted), and
	// manaClassShift its restriction class, 1-based (0 is none).
	manaRestrictedShift = 8
	manaClassShift      = 16
	manaClassMax        = 1<<(32-manaClassShift) - 1
)

// class is u's restriction class: 0 for a unit any cost may spend.
func (u manaUnit) class() int { return int(u >> manaClassShift) }

// restricted is the symbols u pays only for a cost that admits its
// class.
func (u manaUnit) restricted() manaUnit { return (u >> manaRestrictedShift) & manaAny }

// anyMask is every symbol u could pay for some cost.
func (u manaUnit) anyMask() manaUnit { return u&manaAny | u.restricted() }

// payable is the symbols u pays for cost c, with its pain bit: its
// unrestricted symbols, and its restricted ones when c admits its class.
func (u manaUnit) payable(c manaCost) manaUnit {
	m := u & (manaAny | manaPain)
	if c.admitsUnit(u) {
		m |= u.restricted()
	}
	return m
}

// planEpsilon is how much more a plan must be worth than the best
// single move, so a float rounding never replaces today's choice.
const planEpsilon = 1e-9

// manaCost is a cast's mana cost in the model: a generic count and one
// mask per coloured (or {C}) symbol.
type manaCost struct {
	generic int
	colored []manaUnit
	// admits reports whether restricted mana of class k may pay this
	// cost (state.castAdmits). Nil admits none: a filter's input, an
	// activation, an opponent's tax.
	admits func(k int) bool
}

func (c manaCost) total() int { return c.generic + len(c.colored) }

// admitsUnit reports whether u's restricted symbols may pay c.
func (c manaCost) admitsUnit(u manaUnit) bool {
	k := u.class()
	return k > 0 && c.admits != nil && c.admits(k)
}

// manaSource is one activation in the model: the mana it adds, one mask
// per mana, and the mana it consumes first (a filter, a Signet's {1}).
type manaSource struct {
	units []manaUnit
	input int
	// in is the input as a cost: a Signet's {1}, or with
	// Config.PlanFilterLands a filter land's {G/U}, which only its
	// colours can pay. Generic mana of size input when unset.
	in manaCost
	// fallback is what the source makes when the filter cannot be fed:
	// the plain {C} row of a land that has one (Flooded Grove).
	fallback []manaUnit
}

func (s manaSource) net() int { return len(s.units) - s.input }

// planCandidate is one card's best plan-eligible cast in this window.
type planCandidate struct {
	// index is the move's place in this window's list, -1 for a cast
	// the two-turn comparison prices for next turn (rocknow.go).
	index int
	move  legal.Move
	card  *protocol.CardView
	value float64
	// premium is the ramp premium valueOf included, and amount the mana
	// a turn the card adds, which §3 shares across the set.
	premium float64
	amount  int
	cost    manaCost
	// adds is the mana the card makes this turn once it resolves.
	adds []manaSource
	// class is §4's order: 0 adds mana (or, with PlanLandDropsAsRamp,
	// is a permanent that grants extra land drops), 1 draws or tutors,
	// 2 the rest.
	class int
	// baseClass is class without PlanLandDropsAsRamp: the order a set
	// falls back to when the amended order cannot pay for it.
	baseClass int
	// sweep is a declared sweep (ADR 0126 §4). Its price is what it
	// removes from the board as it stands, so among the rest it goes
	// before any permanent the plan adds.
	sweep bool
	// lands is the lands its purpose puts onto the battlefield, net of
	// any it sacrifices (Harrow), which make mana from next turn on
	// (rocknow.go).
	lands int
	// instant is an instant-speed cast whose declared purpose is ramp
	// or draw, which §5 may hold for the end step before the bot's turn
	// (Config.PlanHoldInstants), and drawer one whose purpose draws or
	// tutors.
	instant bool
	drawer  bool
	// annotated is set once rankCandidates has filled the fields above.
	annotated bool
}

// turnPlan is the plan chosen in one window.
type turnPlan struct {
	members []*planCandidate
	value   float64
	// cands are the window's candidates, in §4's order, and mask the
	// members' bits in them.
	cands []*planCandidate
	mask  uint
	// held is the members' bits that §5 holds for the end step before
	// the bot's turn (Config.PlanHoldInstants).
	held uint
}

// planEligible reports whether m may be a plan member (§1, owner answer
// 3): a cast whose only costs are mana and the card itself.
func planEligible(m legal.Move) bool {
	if m.Kind != legal.KindCast {
		return false
	}
	if c := m.Cost; c != nil {
		if c.Life > 0 || c.PhyrexianLife > 0 || c.Loyalty != 0 || len(c.Counters) > 0 || c.Hand > 0 || c.Energy > 0 || c.Exert {
			return false
		}
	}
	return !paramsSet(m.Params, nonManaCastKeys)
}

// planTurn searches the sets of plan-eligible casts on offer for the
// one worth most (§3). vals holds valueOf for each move, already
// priced by decideGeneral. It returns false when no set of two or more
// members is feasible. The plan carries its candidates either way, for
// the two-turn comparison (rocknow.go).
func (p *Policy) planTurn(ctx context.Context, st *state, moves []legal.Move, vals []float64) (turnPlan, bool) {
	cands := p.planCandidates(st, moves, vals)
	out := turnPlan{cands: cands}
	if len(cands) < 2 {
		return out, false
	}
	mask, value, order, held, ok := p.bestSet(ctx, setSearch{
		cands:   cands,
		base:    st.manaAvailable(p.manaModel(st)),
		want:    st.rampWants(),
		sources: st.manaSourcesTotal(),
		minSize: 2,
		hold:    p.cfg.PlanHoldInstants,
	})
	if !ok {
		return out, false
	}
	out.value, out.mask, out.held = value, mask, held
	out.members = membersOf(cands, order, mask, held)
	return out, true
}

// membersOf is the candidates in mask, in order: the ones cast this
// turn first, then the ones held for the end step (§5), in the order
// they are paid.
func membersOf(cands []*planCandidate, order []int, mask, held uint) []*planCandidate {
	var out []*planCandidate
	for _, i := range paySequence(order, mask, held) {
		out = append(out, cands[i])
	}
	return out
}

// setSearch is one search over sets of candidates: this turn's plan
// (§3), or a turn of the two-turn comparison (rocknow.go).
type setSearch struct {
	// cands are in §4's order (planBefore).
	cands []*planCandidate
	// base is the mana there is to spend.
	base []manaUnit
	// want and sources are rampFor's, for the shared premium (§3).
	want    []rampWant
	sources int
	// must is the members every set holds, and minSize the fewest
	// members a set may have.
	must    uint
	minSize int
	// trustSingles takes a set of one as feasible without asking the
	// model: the engine offered the move, and its answer outranks the
	// model's (§2).
	trustSingles bool
	// hold applies §5: an instant-speed member is held for the end step
	// before the bot's turn unless this turn needs it (holdsFor).
	hold bool
}

// bestSet is the feasible set worth most (§3), with the order its
// members are paid in and, with s.hold, the members §5 holds. Ties go
// to the smaller set, then to the set found first. It stops at the
// context deadline with the best set so far.
func (p *Policy) bestSet(ctx context.Context, s setSearch) (uint, float64, []int, uint, bool) {
	cands := s.cands
	maxMana := len(s.base)
	for _, c := range cands {
		for _, a := range c.adds {
			if n := a.net(); n > 0 {
				maxMana += n
			}
		}
	}
	n := len(cands)
	order := identityOrder(n)
	alt, amended := altOrder(cands)
	bestMask, bestVal, bestHeld := uint(0), 0.0, uint(0)
	var bestOrder []int
	units := make([]manaUnit, 0, maxMana)
	for mask := uint(1); mask < 1<<n; mask++ {
		if mask%64 == 0 && ctx.Err() != nil {
			break
		}
		if mask&s.must != s.must {
			continue
		}
		size := bits.OnesCount(mask)
		if size < s.minSize {
			continue
		}
		var cost, amount int
		var value, premium float64
		for i := 0; i < n; i++ {
			if mask&(1<<i) == 0 {
				continue
			}
			c := cands[i]
			cost += c.cost.total()
			value += c.value
			premium += c.premium
			amount += c.amount
		}
		if cost > maxMana {
			continue
		}
		// §3: one shared premium, against the deficit of the cards not
		// in the set, and never more than the members' own premiums.
		value -= premium
		if premium > 0 {
			shared := p.cfg.RampPerMana * float64(min(amount, p.planDeficit(s.want, s.sources, cands, mask)))
			value += min(premium, shared)
		}
		// Ties go to the smaller set (§3), then to the set found first.
		better := bestMask == 0 || value > bestVal+planEpsilon ||
			(value > bestVal-planEpsilon && size < bits.OnesCount(bestMask))
		if !better {
			continue
		}
		used := order
		var held uint
		if size == 1 && s.trustSingles {
			if s.hold {
				held, _ = holdsFor(s.base, cands, order, mask, units)
			}
		} else {
			h, ok := s.feasible(cands, order, mask, units)
			if !ok {
				if mask&amended == 0 {
					continue
				}
				if h, ok = s.feasible(cands, alt, mask, units); !ok {
					continue
				}
				used = alt
			}
			held = h
		}
		bestMask, bestVal, bestOrder, bestHeld = mask, value, used, held
	}
	if bestMask == 0 {
		return 0, 0, nil, 0, false
	}
	return bestMask, bestVal, bestOrder, bestHeld, true
}

// feasible is planFeasible, or with s.hold §5's holds and whether the
// set is payable with them (holdsFor).
func (s setSearch) feasible(cands []*planCandidate, order []int, mask uint, scratch []manaUnit) (uint, bool) {
	if s.hold {
		return holdsFor(s.base, cands, order, mask, scratch)
	}
	return 0, planFeasible(s.base, cands, order, mask, scratch)
}

// identityOrder is 0..n-1: candidates are kept in §4's order.
func identityOrder(n int) []int {
	order := make([]int, n)
	for i := range order {
		order[i] = i
	}
	return order
}

// altOrder is §4's order before the amendment of 2026-10-08, for a set
// whose extra-land-drop member cannot go first and still leave the
// later members payable, and the bits of the candidates the amendment
// moved. Nil and 0 when it moved none.
func altOrder(cands []*planCandidate) ([]int, uint) {
	var amended uint
	for i, c := range cands {
		if c.class != c.baseClass {
			amended |= 1 << i
		}
	}
	if amended == 0 {
		return nil, 0
	}
	alt := identityOrder(len(cands))
	sort.SliceStable(alt, func(i, j int) bool { return planBefore(cands[alt[i]], cands[alt[j]], true) })
	return alt, amended
}

// setFeasible is planFeasible in §4's order, or in the order before its
// amendment when the amended order cannot pay for the set.
func setFeasible(base []manaUnit, cands []*planCandidate, mask uint, scratch []manaUnit) bool {
	if planFeasible(base, cands, identityOrder(len(cands)), mask, scratch) {
		return true
	}
	alt, amended := altOrder(cands)
	return mask&amended != 0 && planFeasible(base, cands, alt, mask, scratch)
}

// planCandidates is each card's best plan-eligible cast, in §4's order,
// the PlanMaxCards most valuable of them.
func (p *Policy) planCandidates(st *state, moves []legal.Move, vals []float64) []*planCandidate {
	out := p.eligibleCasts(st, moves, vals)
	if len(out) < 2 {
		return nil
	}
	return p.rankCandidates(st, out)
}

// eligibleCasts is each card's best plan-eligible cast on offer, in
// the order the moves first name the cards, not yet annotated.
func (p *Policy) eligibleCasts(st *state, moves []legal.Move, vals []float64) []*planCandidate {
	byCard := map[string]*planCandidate{}
	var order []string
	for i := range moves {
		m := moves[i]
		if !p.planEligibleIn(st, m) {
			continue
		}
		cp := decode[castParams](m.Params)
		card := st.castSource(cp.InstanceID)
		if card == nil {
			continue
		}
		if c := byCard[card.InstanceID]; c != nil && c.value >= vals[i] {
			continue
		}
		cost, ok := st.planCastCost(m, card, cp)
		if !ok {
			continue
		}
		c := &planCandidate{index: i, move: m, card: card, value: vals[i], cost: cost}
		if byCard[card.InstanceID] == nil {
			order = append(order, card.InstanceID)
		}
		byCard[card.InstanceID] = c
	}
	out := make([]*planCandidate, 0, len(order))
	for _, id := range order {
		out = append(out, byCard[id])
	}
	return out
}

// rankCandidates keeps the PlanMaxCards most valuable candidates,
// annotates them (premium, amount, adds, class) and sorts them into
// §4's order.
func (p *Policy) rankCandidates(st *state, out []*planCandidate) []*planCandidate {
	sort.SliceStable(out, func(i, j int) bool { return out[i].value > out[j].value })
	if max := p.cfg.PlanMaxCards; max > 0 && len(out) > max {
		out = out[:max]
	}
	// What each candidate's price owes to the ramp premium, and what it
	// adds this turn. Priced again with the premium off: the difference
	// is exactly what the premium added after SpellFloor.
	noRamp := &Policy{cfg: p.cfg}
	noRamp.cfg.RampPerMana = 0
	for _, c := range out {
		if c.annotated {
			continue
		}
		c.annotated = true
		m := c.move
		cp := decode[castParams](m.Params)
		ps := castPurpose(c.card, cp)
		if p.cfg.RampPerMana > 0 {
			if v, _ := noRamp.valueOf(st, m); c.value > v {
				c.premium = c.value - v
			}
			if !isLand(c.card) {
				c.amount += repeatableMana(c.card)
			}
			if p.cfg.PricePurposes {
				c.amount += max(0, ps.lands-len(cp.SacrificeIDs))
			}
		}
		if p.cfg.PricePurposes {
			c.lands = max(0, ps.lands-len(cp.SacrificeIDs))
		}
		c.adds = castAddsMana(c.card, ps, p.manaModel(st))
		c.drawer = ps.draws > 0 || ps.tutors > 0
		// The owner's decision of 2026-10-09 (ADR 0136 §5, question 2):
		// only an instant-speed cast whose declared purpose is ramp or
		// draw is held. Removal, burn and flash creatures are not.
		ramp := ps.lands > 0 || len(c.adds) > 0 || (!isLand(c.card) && repeatableMana(c.card) > 0)
		c.instant = p.cfg.PlanHoldInstants && p.instantSpeed(st, m) && (ramp || c.drawer)
		switch {
		case len(c.adds) > 0:
			c.baseClass = 0
		case ps.draws > 0 || ps.tutors > 0:
			c.baseClass = 1
		default:
			c.baseClass = 2
		}
		c.class = c.baseClass
		if p.cfg.PlanLandDropsAsRamp && c.class > 0 && ps.extraLands > 0 && isPermanentSpell(c.card) && !isLand(c.card) {
			// The amendment of 2026-10-08: a permanent that grants extra
			// land drops is ramp in §4's order, cast before the draws so
			// a land they find can still be played this turn.
			c.class = 0
		}
		c.sweep = p.cfg.PriceSweeps && len(ps.sweeps) > 0
	}
	// §4's order, so a set's members are always taken in it.
	sort.SliceStable(out, func(i, j int) bool { return planBefore(out[i], out[j], false) })
	return out
}

// planBefore is §4's order: mana first, then draws and tutors, the
// cheapest first in each, then the rest with a declared sweep first and
// then the highest value. base orders by each candidate's class before
// the amendment of 2026-10-08 (planCandidate.baseClass).
func planBefore(a, b *planCandidate, base bool) bool {
	ac, bc := a.class, b.class
	if base {
		ac, bc = a.baseClass, b.baseClass
	}
	if ac != bc {
		return ac < bc
	}
	if ac < 2 && a.cost.total() != b.cost.total() {
		return a.cost.total() < b.cost.total()
	}
	if a.sweep != b.sweep {
		return a.sweep
	}
	if a.value != b.value {
		return a.value > b.value
	}
	return a.index < b.index
}

// planFeasible reports whether the members in mask can be paid in
// order (indices into cands: §4's order, or the order before its
// amendment) from base, each member's mana arriving for the ones after
// it. scratch is reused between calls.
//
// A set holding a permanent that grants extra land drops is tried in
// the amended order first. Only if that order cannot pay for it is the
// set tried in the order before the amendment (the draw first): the
// amendment changes which member goes first, never which sets the bot
// can afford.
func planFeasible(base []manaUnit, cands []*planCandidate, order []int, mask uint, scratch []manaUnit) bool {
	_, ok := planWalk(base, cands, order, mask, 0, -1, scratch)
	return ok
}

// planWalk pays the members in mask in order from base, each member's
// mana arriving for the ones after it, and reports whether every one
// was paid. The members in held (§5) are paid last, after every member
// cast this turn, in the end step before the bot's turn: their mana
// stays reserved, and what they add arrives only for each other. left
// is the mana there is right after the candidate at index `after` is
// paid and its mana added, -1 when it is not reached.
func planWalk(base []manaUnit, cands []*planCandidate, order []int, mask, held uint, after int, scratch []manaUnit) (left int, ok bool) {
	units := append(scratch[:0], base...)
	seq := paySequence(order, mask, held)
	left = -1
	for k, i := range seq {
		c := cands[i]
		// The coloured symbols of the members still to come: generic
		// mana is paid around them.
		var later []manaUnit
		for _, j := range seq[k+1:] {
			later = append(later, cands[j].cost.colored...)
		}
		if units, ok = payMana(units, c.cost, later); !ok {
			return left, false
		}
		for _, a := range c.adds {
			units = addSource(units, a, later)
		}
		if i == after {
			left = len(units)
		}
	}
	return left, true
}

// paySequence is the members in mask in order, the ones in held moved
// after the rest.
func paySequence(order []int, mask, held uint) []int {
	seq := make([]int, 0, bits.OnesCount(mask))
	for pass := 0; pass < 2; pass++ {
		for _, i := range order {
			bit := uint(1) << i
			if mask&bit == 0 || (held&bit != 0) != (pass == 1) {
				continue
			}
			seq = append(seq, i)
		}
	}
	return seq
}

// planDeficit is rampFor's deficit with every card in the set left out
// of `want` (§3): a rock whose only purpose was to reach a spell the
// plan casts anyway closes nothing.
func (p *Policy) planDeficit(want []rampWant, sources int, cands []*planCandidate, mask uint) int {
	w := 0
	for _, rw := range want {
		in := false
		for i, c := range cands {
			if mask&(1<<i) != 0 && c.card.InstanceID == rw.id {
				in = true
				break
			}
		}
		if !in {
			w = rw.mv
			break
		}
	}
	if w > p.cfg.RampWantCap {
		w = p.cfg.RampWantCap
	}
	return max(0, w-sources)
}

// rampWant is one card's claim on the ramp deficit: its mana value, the
// commander's with its tax.
type rampWant struct {
	id string
	mv int
}

// rampWants is rampFor's `want` candidates, largest first.
func (st *state) rampWants() []rampWant {
	if st.seat == nil {
		return nil
	}
	var out []rampWant
	for i := range st.seat.Hand.Cards {
		c := &st.seat.Hand.Cards[i]
		if !isLand(c) {
			out = append(out, rampWant{c.InstanceID, manaValue(c.ManaCost, 0)})
		}
	}
	for i := range st.seat.Command.Cards {
		c := &st.seat.Command.Cards[i]
		if !isLand(c) {
			out = append(out, rampWant{c.InstanceID, manaValue(c.ManaCost, 0) + 2*st.seat.CommanderCasts[c.InstanceID]})
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].mv > out[j].mv })
	return out
}

// manaSourcesTotal is rampFor's `sources`: the mana the bot's own
// sources make, tapped or not.
func (st *state) manaSourcesTotal() int {
	n := 0
	for _, c := range st.bf {
		if c.Controller == st.me {
			n += repeatableMana(c)
		}
	}
	return n
}

// planCastCost is the mana a cast move charges (§2): the move's own
// total (legal.MoveCost.Mana, CR 601.2f) when the enumerator stamped
// it. A view from before that field is priced from the printed cost,
// the announced X and the commander tax (CR 903.8); an alternative cost
// it cannot read keeps the move out of the plan.
//
// Restricted mana pays it when its restriction admits this cast
// (castAdmits).
func (st *state) planCastCost(m legal.Move, card *protocol.CardView, cp castParams) (manaCost, bool) {
	if m.Cost != nil && m.Cost.Mana != "" {
		c := parseManaCost(m.Cost.Mana, 0)
		c.admits = st.castAdmits(card, cp.FromZone)
		return c, true
	}
	if cp.AlternativeCost != "" || card.ManaCost == "" {
		return manaCost{}, false
	}
	c := parseManaCost(card.ManaCost, cp.XValue)
	if cp.FromZone == "command" && st.seat != nil {
		c.generic += 2 * st.seat.CommanderCasts[card.InstanceID]
	}
	c.admits = st.castAdmits(card, cp.FromZone)
	return c, true
}

// manaModel is how the plan reads the bot's mana sources (§2).
type manaModel struct {
	// filterLands is Config.PlanFilterLands: a filter land is its
	// filter (manaSourceOf).
	filterLands bool
	// classes, with Config.PlanRestrictedMana, numbers the restricted
	// rows the model counts. Nil leaves every restricted row out.
	classes *manaClasses
}

// manaModel is the mana model p plans with in st.
func (p *Policy) manaModel(st *state) manaModel {
	mm := manaModel{filterLands: p.cfg.PlanFilterLands}
	if p.cfg.PlanRestrictedMana {
		mm.classes = st.restrictionClasses()
	}
	return mm
}

// manaClasses numbers the distinct restriction lists of the restricted
// rows the model counts (ADR 0136's amendment of 2026-10-16), so a unit
// can carry its row's list as one small number (manaUnit.class). Class
// k is lists[k-1].
type manaClasses struct {
	lists [][]string
	index map[string]int
}

// classOf is the class of restriction list rs, numbered on first sight;
// 0 when there is no room for another.
func (r *manaClasses) classOf(rs []string) int {
	key := strings.Join(rs, "\x00")
	if k, ok := r.index[key]; ok {
		return k
	}
	if len(r.lists) >= manaClassMax {
		return 0
	}
	if r.index == nil {
		r.index = map[string]int{}
	}
	r.lists = append(r.lists, rs)
	r.index[key] = len(r.lists)
	return len(r.lists)
}

// restrictionClasses is this window's class numbering, one for the whole
// decision, so every unit and every cost agree on it.
func (st *state) restrictionClasses() *manaClasses {
	if st.restrictClasses == nil {
		st.restrictClasses = &manaClasses{}
	}
	return st.restrictClasses
}

// castAdmits is a cast's manaCost.admits: whether class k's restriction
// list admits casting card from fromZone, asked of the engine's matcher
// (protocol.ManaRestrictionsAdmitCast, the one auto-pay plans by) once
// per class.
func (st *state) castAdmits(card *protocol.CardView, fromZone string) func(k int) bool {
	memo := map[int]bool{}
	return func(k int) bool {
		if v, ok := memo[k]; ok {
			return v
		}
		r := st.restrictClasses
		v := r != nil && k >= 1 && k <= len(r.lists) && protocol.ManaRestrictionsAdmitCast(r.lists[k-1], card, fromZone)
		memo[k] = v
		return v
	}
}

// parseManaCost reads "{5}{G}{U}" into the model. X is x. A hybrid
// symbol is paid by either half, a Phyrexian symbol by its colour (its
// life is already on the move's Life when it is paid that way), and a
// {2/W} by W.
func parseManaCost(s string, x int) manaCost {
	var c manaCost
	for _, sym := range manaSymbols(s) {
		if n, err := strconv.Atoi(sym); err == nil {
			c.generic += n
			continue
		}
		switch sym {
		case "X", "Y", "Z":
			c.generic += x
			continue
		case "S":
			c.colored = append(c.colored, manaAny)
			continue
		}
		var mask manaUnit
		for _, part := range strings.Split(sym, "/") {
			mask |= colorMask(part)
		}
		if mask == 0 {
			c.generic++
			continue
		}
		c.colored = append(c.colored, mask)
	}
	return c
}

// colorMask is one mana letter's mask, 0 for anything else.
func colorMask(letter string) manaUnit {
	switch letter {
	case "W":
		return manaW
	case "U":
		return manaU
	case "B":
		return manaB
	case "R":
		return manaR
	case "G":
		return manaG
	case "C":
		return manaC
	}
	return 0
}

// manaAvailable is the mana the bot can make now (§2): its pool, and
// each untapped permanent it controls with a repeatable mana ability it
// can activate, filters already run, read under mm (manaSourceOf).
func (st *state) manaAvailable(mm manaModel) []manaUnit {
	var units []manaUnit
	if st.seat != nil {
		for _, run := range st.seat.ManaPool {
			if m := colorMask(run.Color); m != 0 {
				for range run.Count {
					units = append(units, m)
				}
			}
		}
	}
	var filters []manaSource
	for i := range st.view.Battlefield.Cards {
		c := &st.view.Battlefield.Cards[i]
		if c.Controller != st.me || c.Tapped || (isCreature(c) && c.ActivationSick()) {
			continue
		}
		src, ok := manaSourceOf(c, mm)
		if !ok {
			continue
		}
		if src.input > 0 {
			filters = append(filters, src)
			continue
		}
		units = append(units, src.units...)
	}
	for _, f := range filters {
		units = addSource(units, f, nil)
	}
	return units
}

// manaSourceOf is c's best repeatable mana ability as a source (§2):
// a {T} ability with no sacrifice, exile or other non-mana cost that is
// not greyed out. A life cost or a damage rider still counts. Rows that
// net the same mana are merged, so a painland is one source of either
// colour, and a filter row loses to a plain one that nets as much,
// unless mm.filterLands is set and the plain row makes only colourless
// mana: then the filter is the source (§2: "a filter ability (a Signet,
// Flooded Grove) is an entry that consumes one mana and makes its
// output"), its input paid with the colours it names, and the plain
// row is what it makes when nothing can feed it. A
// land with no mana rows (a fetchland) makes no mana: every land that
// taps for mana carries a row, its intrinsic ones included.
//
// A row whose mana may be spent only on some spells (Delighted
// Halfling's colours, Cavern of Souls' chosen type, Ancient Ziggurat) is
// counted only with mm.classes (Config.PlanRestrictedMana), and then
// only for the casts its restriction admits (restrictedOverlay). Without
// it the row is left out.
func manaSourceOf(c *protocol.CardView, mm manaModel) (manaSource, bool) {
	var best, filter manaSource
	found, haveFilter := false, false
	var rest []manaUnit
	var restClass int
	for i := range c.ManaAbilities {
		ab := &c.ManaAbilities[i]
		if !ab.TapCost || ab.SacrificeCost || ab.ExileSelf || ab.AddsNoMana || ab.ConditionUnmet ||
			ab.Exhausted || ab.CantActivate != "" || ab.SacrificeLabel != "" || ab.TapOthersLabel != "" ||
			ab.ExilePermanentLabel != "" || ab.Exert || ab.EnergyCost > 0 || ab.DiscardCostN > 0 ||
			ab.ExileCostN > 0 || ab.CounterCostN > 0 || ab.CounterCostVariable {
			continue
		}
		cost := ab.ManaCost
		if ab.ChargedManaCost != nil {
			cost = *ab.ChargedManaCost
		}
		if len(ab.Restrictions) > 0 {
			// The restricted row as mana a cast it admits may spend:
			// painless and costing nothing but the tap, so that each
			// unit is one mana of it. The largest such row is kept.
			if mm.classes == nil || manaValue(cost, 0) > 0 || ab.LifeCost > 0 ||
				strings.Contains(strings.ToLower(ab.Label), "damage to you") {
				continue
			}
			if units := abilityUnits(ab); len(units) > len(rest) {
				if k := mm.classes.classOf(ab.Restrictions); k > 0 {
					rest, restClass = units, k
				}
			}
			continue
		}
		src := manaSource{units: abilityUnits(ab), input: manaValue(cost, 0)}
		src.in = manaCost{generic: src.input}
		if ab.LifeCost > 0 || strings.Contains(strings.ToLower(ab.Label), "damage to you") {
			for k := range src.units {
				src.units[k] |= manaPain
			}
		}
		if mm.filterLands && src.input > 0 && colouredOnly(src.units) && (!haveFilter || src.net() > filter.net()) {
			filter, haveFilter = src, true
			filter.in = parseManaCost(cost, 0)
		}
		switch {
		case !found || src.net() > best.net() || (src.net() == best.net() && src.input < best.input):
			best, found = src, true
		case src.net() == best.net() && src.input == best.input && len(src.units) == len(best.units):
			// Either row: the colours of both, and painless if either is
			// (a painland's {C} is).
			for k := range best.units {
				pain := best.units[k] & src.units[k] & manaPain
				best.units[k] = (best.units[k]|src.units[k])&^manaPain | pain
			}
		}
	}
	if haveFilter && found && best.input == 0 && filter.net() == best.net() && !colouredAny(best.units) {
		filter.fallback = best.units
		return filter, true
	}
	if rest != nil && (!found || best.input == 0) {
		var plain []manaUnit
		if found {
			plain = best.units
		}
		if units, ok := restrictedOverlay(plain, rest, restClass); ok {
			return manaSource{units: units}, true
		}
	}
	return best, found && best.net() > 0
}

// restrictedOverlay is one activation of a source with a plain row
// (plain, nil when it has none) and a restricted row (rest, of class k)
// as units (ADR 0136's amendment of 2026-10-16): unit i pays plain[i]
// for any cost, and rest[i] too for a cost that admits k. It is exact
// when no mix of the two rows can be spent that neither row makes: no
// plain row (Ancient Ziggurat), one mana from each (Delighted
// Halfling's {C} or a colour, Cavern of Souls'), or a restricted row
// that makes at least as much, each mana of it paying what the plain
// row's does (Eldrazi Temple's {C} or {C}{C}). Any other pair is not
// overlaid (false), and the plain row alone is the source, as it was
// before restricted rows were counted.
func restrictedOverlay(plain, rest []manaUnit, k int) ([]manaUnit, bool) {
	exact := len(plain) == 0 || (len(plain) == 1 && len(rest) == 1)
	if !exact && len(rest) >= len(plain) {
		exact = true
		for i, u := range plain {
			if u&manaPain != 0 || u&manaAny&^rest[i] != 0 {
				exact = false
				break
			}
		}
	}
	if !exact {
		return nil, false
	}
	out := make([]manaUnit, max(len(plain), len(rest)))
	for i := range out {
		if i < len(plain) {
			out[i] = plain[i]
		}
		if i < len(rest) {
			out[i] |= (rest[i]&manaAny)<<manaRestrictedShift | manaUnit(k)<<manaClassShift
		}
	}
	return out, true
}

// colouredOnly reports whether every unit is a painless coloured mana.
func colouredOnly(units []manaUnit) bool {
	for _, u := range units {
		if u&manaAnyColor == 0 || u&manaPain != 0 {
			return false
		}
	}
	return len(units) > 0
}

// colouredAny reports whether any unit can be coloured mana.
func colouredAny(units []manaUnit) bool {
	for _, u := range units {
		if u&manaAnyColor != 0 {
			return true
		}
	}
	return false
}

// abilityUnits is the mana one activation adds, one mask per mana: from
// `produced`, with each choice narrowed by its color_options slot where
// the view stamps one (Command Tower, Birds of Paradise). An output the
// view cannot size (Exotic Orchard, Cabal Coffers) is one mana of the
// colours color_options names, or of no colour.
func abilityUnits(ab *protocol.ManaAbilityView) []manaUnit {
	slot := 0
	optionMask := func() manaUnit {
		if slot >= len(ab.ColorOptions) {
			return 0
		}
		var m manaUnit
		for _, l := range ab.ColorOptions[slot] {
			m |= colorMask(l)
		}
		slot++
		return m
	}
	if ab.Produced == "" {
		if m := optionMask(); m != 0 {
			return []manaUnit{m}
		}
		return []manaUnit{manaC}
	}
	var units []manaUnit
	for _, sym := range manaSymbols(ab.Produced) {
		// #2558: "{W|U|B|R|G:2}" is two mana of different colors — one
		// unit per mana, each from its own color_options list. A unit
		// mask cannot say "different", so two such units read as one
		// pair of any colours, which the server's move list corrects.
		if k := strings.LastIndexByte(sym, ':'); k >= 0 {
			if v, err := strconv.Atoi(sym[k+1:]); err == nil && v > 0 {
				var mask manaUnit
				for _, opt := range strings.Split(sym[:k], "|") {
					mask |= colorMask(opt)
				}
				for i := 0; i < v; i++ {
					u := mask
					if m := optionMask(); m != 0 {
						u = m
					}
					units = append(units, u)
				}
				continue
			}
		}
		n := 1
		var mask manaUnit
		choice := strings.IndexByte(sym, '|') >= 0
		for _, opt := range strings.Split(sym, "|") {
			letter := opt
			if k := strings.IndexAny(opt, "0123456789"); k > 0 {
				if v, err := strconv.Atoi(opt[k:]); err == nil && v > 0 {
					n = v
				}
				letter = opt[:k]
			} else if v, err := strconv.Atoi(opt); err == nil {
				// A bare number is that much colourless.
				n, letter = v, "C"
			}
			mask |= colorMask(letter)
		}
		if choice {
			if m := optionMask(); m != 0 {
				mask = m
			}
		}
		if mask == 0 {
			mask = manaC
		}
		for k := 0; k < n; k++ {
			units = append(units, mask)
		}
	}
	if len(units) == 0 {
		return []manaUnit{manaC}
	}
	return units
}

// castAddsMana is the mana a card adds THIS turn once it resolves (§2):
// a noncreature permanent's repeatable mana ability (a rock enters
// untapped and has no summoning sickness, CR 302.6), a hasty creature's,
// and one mana of any colour for each land its purpose says enters
// untapped (PurposeView.lands_untapped, owner answer 4).
func castAddsMana(c *protocol.CardView, ps purposeSet, mm manaModel) []manaSource {
	var out []manaSource
	if !isLand(c) && isPermanentSpell(c) && (!isCreature(c) || hasKeyword(c, "haste")) {
		if src, ok := manaSourceOf(c, mm); ok && len(c.ManaAbilities) > 0 {
			out = append(out, src)
		}
	}
	for k := 0; k < ps.landsUntapped; k++ {
		out = append(out, manaSource{units: []manaUnit{manaAnyColor}})
	}
	return out
}

// payMana pays cost out of units (§2) the way the engine's auto-tapper
// would, since that is what will pay it: coloured symbols first, the
// most constrained symbol first, each from the painless and least
// flexible unit that can pay it; then generic from painless before
// painful, colourless before coloured, fewest colours first
// (engineOrder). The auto-tapper does not know what the plan casts
// next, so where its order ties the model assumes the worst: it spends
// the unit the later members' symbols (later) need most. A plan that
// survives that is one the engine will let the bot finish. It reorders
// units in place.
//
// Restricted mana (manaUnit.class) pays only a cost that admits it, and
// goes first there (spendBefore). A unit the cost cannot spend at all
// (restricted mana with no plain row, for a cast its restriction
// refuses) is set aside and kept for the members after it.
func payMana(units []manaUnit, cost manaCost, later []manaUnit) ([]manaUnit, bool) {
	var aside []manaUnit
	if unusable(units, cost) {
		usable := make([]manaUnit, 0, len(units))
		for _, u := range units {
			if u.payable(cost)&manaAny != 0 {
				usable = append(usable, u)
			} else {
				aside = append(aside, u)
			}
		}
		units = usable
	}
	if cost.total() > len(units) {
		return append(units, aside...), false
	}
	colored := append([]manaUnit(nil), cost.colored...)
	// The most constrained symbol first: fewest units can pay it.
	sort.SliceStable(colored, func(i, j int) bool {
		return payers(units, colored[i], cost) < payers(units, colored[j], cost)
	})
	for _, sym := range colored {
		pick := -1
		for k, u := range units {
			if u.payable(cost)&sym == 0 {
				continue
			}
			if pick < 0 || spendBefore(u, units[pick], later, cost) {
				pick = k
			}
		}
		if pick < 0 {
			return append(units, aside...), false
		}
		units = removeUnit(units, pick)
	}
	if cost.generic > len(units) {
		return append(units, aside...), false
	}
	engineOrder(units, later, cost)
	return append(units[cost.generic:], aside...), true
}

// unusable reports whether any unit can pay nothing of cost: restricted
// mana with no plain row, which cost does not admit.
func unusable(units []manaUnit, cost manaCost) bool {
	for _, u := range units {
		if u.class() != 0 && u.payable(cost)&manaAny == 0 {
			return true
		}
	}
	return false
}

// engineOrder sorts units into the order the auto-tapper spends them on
// cost's generic mana, worst case first within its ties (payMana).
func engineOrder(units, later []manaUnit, cost manaCost) {
	sort.SliceStable(units, func(i, j int) bool { return spendBefore(units[i], units[j], later, cost) })
}

// spendBefore reports whether the auto-tapper spends a before b on cost:
// painless first, then restricted mana the cost may spend (ADR 0040's
// amendment of 2026-10-16 ranks it, and the rider it fires, ahead of
// mana any cast could spend), then the fewest colours (colourless is
// none), and within a tie the one the later symbols need more.
func spendBefore(a, b manaUnit, later []manaUnit, cost manaCost) bool {
	pa, pb := a.payable(cost), b.payable(cost)
	if ap, bp := pa&manaPain != 0, pb&manaPain != 0; ap != bp {
		return !ap
	}
	if ra, rb := cost.admitsUnit(a), cost.admitsUnit(b); ra != rb {
		return ra
	}
	if ca, cb := bits.OnesCount32(uint32(pa&manaAnyColor)), bits.OnesCount32(uint32(pb&manaAnyColor)); ca != cb {
		return ca < cb
	}
	return demand(a, later) > demand(b, later)
}

// demand is how many of the later symbols u could pay, its restricted
// mana counted as if a later cost admits it: the worse case.
func demand(u manaUnit, later []manaUnit) int {
	n := 0
	m := u.anyMask()
	for _, sym := range later {
		if m&sym != 0 {
			n++
		}
	}
	return n
}

// payers is how many units can pay sym for cost.
func payers(units []manaUnit, sym manaUnit, cost manaCost) int {
	n := 0
	for _, u := range units {
		if u.payable(cost)&sym != 0 {
			n++
		}
	}
	return n
}

// manaSymbols splits "{2}{G}" into "2", "G".
func manaSymbols(s string) []string {
	var out []string
	for {
		i := strings.IndexByte(s, '{')
		if i < 0 {
			return out
		}
		j := strings.IndexByte(s[i:], '}')
		if j < 0 {
			return out
		}
		if sym := s[i+1 : i+j]; sym != "" {
			out = append(out, sym)
		}
		s = s[i+j+1:]
	}
}

func removeUnit(units []manaUnit, k int) []manaUnit {
	units[k] = units[len(units)-1]
	return units[:len(units)-1]
}

// addSource activates src: a filter consumes its input as generic mana
// is paid (engineOrder), and is not run when there is not enough to
// feed it. A filter whose input names colours (PlanFilterLands) pays it
// as a cast would (payMana), and makes its fallback when it cannot.
func addSource(units []manaUnit, src manaSource, later []manaUnit) []manaUnit {
	if len(src.in.colored) > 0 {
		rest, ok := payMana(append([]manaUnit(nil), units...), src.in, later)
		if !ok {
			return append(units, src.fallback...)
		}
		return append(rest, src.units...)
	}
	if src.input > 0 {
		// Mana a cast's restriction admits does not feed a filter: its
		// input is an activation's cost (manaCost.admits is nil).
		rest, ok := payMana(units, manaCost{generic: src.input}, later)
		if !ok {
			return rest
		}
		units = rest
	}
	return append(units, src.units...)
}

// planMembers is the plan as the trace records it (§7), held members
// marked (§5).
func planMembers(moves []legal.Move, pl turnPlan) []aiseat.PlanMember {
	out := make([]aiseat.PlanMember, 0, len(pl.members))
	for _, c := range pl.members {
		out = append(out, aiseat.PlanMember{Index: c.index, Label: moves[c.index].Label, Held: pl.isHeld(c)})
	}
	return out
}

// isHeld reports whether member c is held for the end step (§5).
func (pl turnPlan) isHeld(c *planCandidate) bool {
	for i, k := range pl.cands {
		if k == c {
			return pl.held&(1<<i) != 0
		}
	}
	return false
}

// planReason is the decision's reason (§7): "plan: Arcane Signet →
// Ornithopter of Paradise (+2.76)", a held member named as "Harrow at
// the end step" (§5).
func planReason(pl turnPlan) string {
	names := make([]string, 0, len(pl.members))
	for _, c := range pl.members {
		if pl.isHeld(c) {
			names = append(names, c.card.Name+" at the end step")
			continue
		}
		names = append(names, c.card.Name)
	}
	return fmt.Sprintf("plan: %s (+%.2f)", strings.Join(names, " → "), pl.value)
}

// decidePlan is the plan's half of decideGeneral (§3): in the bot's own
// main phase with an empty stack, when the best set of two or more
// casts is worth more than the best single move and clears the
// window's bar, make the set's first move. best and bestVal are the
// best single move; threshold is the window's bar, and leftover says
// ADR 0126 §5's lower bar applies to a plan whose every member is
// eligible for it.
//
// The land drop is left alone (§1): while a land is the best move the
// bot plays it, and plans the rest in the next window.
//
// The plan it searched comes back whether or not it was chosen, with
// its candidates, for the two-turn comparison (rocknow.go). searched
// is false when no search ran.
func (p *Policy) decidePlan(ctx context.Context, st *state, moves []legal.Move, vals []float64,
	best int, bestVal, threshold float64, leftover bool) (d aiseat.Decision, members []aiseat.PlanMember, pl turnPlan, searched, chosen bool) {
	if !p.cfg.PlanTurnMana || !st.sorcerySpeed || best < 0 || moves[best].Kind == legal.KindLand || ctx.Err() != nil {
		return aiseat.Decision{}, nil, turnPlan{}, false, false
	}
	pl, ok := p.planTurn(ctx, st, moves, vals)
	if !ok || len(pl.members) < 2 || pl.value <= bestVal+planEpsilon {
		return aiseat.Decision{}, nil, pl, true, false
	}
	if pl.value <= p.setBar(st, moves, pl.members, threshold, leftover) {
		return aiseat.Decision{}, nil, pl, true, false
	}
	// §5: the first member cast this turn. The held members come last,
	// so when the first is held every member is, and the bot passes.
	d = aiseat.Decision{Index: pl.members[0].index, Reason: planReason(pl)}
	if pl.isHeld(pl.members[0]) {
		d.Index = passOrDecline(moves)
	}
	return d, planMembers(moves, pl), pl, true, true
}

// setBar is the bar a set of this turn's casts clears: the window's,
// or ADR 0126 §5's LeftoverThreshold when the window is a leftover one
// and every member is eligible for it.
func (p *Policy) setBar(st *state, moves []legal.Move, members []*planCandidate, threshold float64, leftover bool) float64 {
	if !leftover {
		return threshold
	}
	for _, c := range members {
		if !p.leftoverEligible(st, moves[c.index]) {
			return threshold
		}
	}
	return p.cfg.LeftoverThreshold
}
