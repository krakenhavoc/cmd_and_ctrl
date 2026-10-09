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
//     are mana and the card. Each card contributes its best such move.
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
// Instants are cast like sorceries here. Holding one for the end step
// before the bot's turn (§5) is ADR 0136 PR 5.

// Colour masks for one mana in the model.
const (
	manaW uint8 = 1 << iota
	manaU
	manaB
	manaR
	manaG
	manaC

	// manaPain marks a mana whose source costs life or deals damage to
	// its controller (Mana Confluence, Ancient Tomb, City of Brass). It
	// is not a colour: the engine's auto-tapper spends such a source
	// only once nothing painless is left, and the model does the same.
	manaPain uint8 = 1 << 7

	manaAnyColor = manaW | manaU | manaB | manaR | manaG
	// manaAny pays any symbol at all, {C} included (a snow symbol's
	// stand-in).
	manaAny = manaAnyColor | manaC
)

// planEpsilon is how much more a plan must be worth than the best
// single move, so a float rounding never replaces today's choice.
const planEpsilon = 1e-9

// manaCost is a cast's mana cost in the model: a generic count and one
// mask per coloured (or {C}) symbol.
type manaCost struct {
	generic int
	colored []uint8
}

func (c manaCost) total() int { return c.generic + len(c.colored) }

// manaSource is one activation in the model: the mana it adds, one mask
// per mana, and the mana it consumes first (a filter, a Signet's {1}).
type manaSource struct {
	units []uint8
	input int
	// in is the input as a cost: a Signet's {1}, or with
	// Config.PlanFilterLands a filter land's {G/U}, which only its
	// colours can pay. Generic mana of size input when unset.
	in manaCost
	// fallback is what the source makes when the filter cannot be fed:
	// the plain {C} row of a land that has one (Flooded Grove).
	fallback []uint8
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
	// lands is the lands its purpose puts onto the battlefield, which
	// make mana from next turn on (rocknow.go).
	lands int
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
	mask, value, order, ok := p.bestSet(ctx, setSearch{
		cands:   cands,
		base:    st.manaAvailable(p.cfg.PlanFilterLands),
		want:    st.rampWants(),
		sources: st.manaSourcesTotal(),
		minSize: 2,
	})
	if !ok {
		return out, false
	}
	out.value, out.mask = value, mask
	out.members = membersOf(cands, order, mask)
	return out, true
}

// membersOf is the candidates in mask, in order.
func membersOf(cands []*planCandidate, order []int, mask uint) []*planCandidate {
	var out []*planCandidate
	for _, i := range order {
		if mask&(1<<i) != 0 {
			out = append(out, cands[i])
		}
	}
	return out
}

// setSearch is one search over sets of candidates: this turn's plan
// (§3), or a turn of the two-turn comparison (rocknow.go).
type setSearch struct {
	// cands are in §4's order (planBefore).
	cands []*planCandidate
	// base is the mana there is to spend.
	base []uint8
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
}

// bestSet is the feasible set worth most (§3), with the order its
// members are paid in. Ties go to the smaller set, then to the set
// found first. It stops at the context deadline with the best set so
// far.
func (p *Policy) bestSet(ctx context.Context, s setSearch) (uint, float64, []int, bool) {
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
	bestMask, bestVal := uint(0), 0.0
	var bestOrder []int
	units := make([]uint8, 0, maxMana)
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
		if !(size == 1 && s.trustSingles) && !planFeasible(s.base, cands, order, mask, units) {
			if mask&amended == 0 || !planFeasible(s.base, cands, alt, mask, units) {
				continue
			}
			used = alt
		}
		bestMask, bestVal, bestOrder = mask, value, used
	}
	if bestMask == 0 {
		return 0, 0, nil, false
	}
	return bestMask, bestVal, bestOrder, true
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
func setFeasible(base []uint8, cands []*planCandidate, mask uint, scratch []uint8) bool {
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
		if !planEligible(m) {
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
				c.amount += ps.lands
			}
		}
		if p.cfg.PricePurposes {
			c.lands = ps.lands
		}
		c.adds = castAddsMana(c.card, ps, p.cfg.PlanFilterLands)
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
func planFeasible(base []uint8, cands []*planCandidate, order []int, mask uint, scratch []uint8) bool {
	units := append(scratch[:0], base...)
	for k, i := range order {
		if mask&(1<<i) == 0 {
			continue
		}
		c := cands[i]
		// The coloured symbols of the members still to come: generic
		// mana is paid around them.
		var later []uint8
		for _, j := range order[k+1:] {
			if mask&(1<<j) != 0 {
				later = append(later, cands[j].cost.colored...)
			}
		}
		var ok bool
		if units, ok = payMana(units, c.cost, later); !ok {
			return false
		}
		for _, a := range c.adds {
			units = addSource(units, a, later)
		}
	}
	return true
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
func (st *state) planCastCost(m legal.Move, card *protocol.CardView, cp castParams) (manaCost, bool) {
	if m.Cost != nil && m.Cost.Mana != "" {
		return parseManaCost(m.Cost.Mana, 0), true
	}
	if cp.AlternativeCost != "" || card.ManaCost == "" {
		return manaCost{}, false
	}
	c := parseManaCost(card.ManaCost, cp.XValue)
	if cp.FromZone == "command" && st.seat != nil {
		c.generic += 2 * st.seat.CommanderCasts[card.InstanceID]
	}
	return c, true
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
		var mask uint8
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
func colorMask(letter string) uint8 {
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
// can activate, filters already run. With filterLands, a filter land is
// its filter (manaSourceOf).
func (st *state) manaAvailable(filterLands bool) []uint8 {
	var units []uint8
	if st.seat != nil {
		for _, sym := range st.seat.ManaPool {
			if m := colorMask(sym); m != 0 {
				units = append(units, m)
			}
		}
	}
	var filters []manaSource
	for i := range st.view.Battlefield.Cards {
		c := &st.view.Battlefield.Cards[i]
		if c.Controller != st.me || c.Tapped || (isCreature(c) && c.SummoningSick) {
			continue
		}
		src, ok := manaSourceOf(c, filterLands)
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
// unless filterLands is set and the plain row makes only colourless
// mana: then the filter is the source (§2: "a filter ability (a Signet,
// Flooded Grove) is an entry that consumes one mana and makes its
// output"), its input paid with the colours it names, and the plain
// row is what it makes when nothing can feed it. A
// row whose mana may be spent only on some spells (Delighted Halfling's
// colours, Ancient Ziggurat) is left out: the model cannot tell which
// casts it pays for. A
// land with no mana rows (a fetchland) makes no mana: every land that
// taps for mana carries a row, its intrinsic ones included.
func manaSourceOf(c *protocol.CardView, filterLands bool) (manaSource, bool) {
	var best, filter manaSource
	found, haveFilter := false, false
	for i := range c.ManaAbilities {
		ab := &c.ManaAbilities[i]
		if !ab.TapCost || ab.SacrificeCost || ab.ExileSelf || ab.AddsNoMana || ab.ConditionUnmet ||
			ab.Exhausted || ab.CantActivate != "" || ab.SacrificeLabel != "" || ab.TapOthersLabel != "" ||
			ab.ExilePermanentLabel != "" || ab.Exert || ab.EnergyCost > 0 || ab.DiscardCostN > 0 ||
			ab.ExileCostN > 0 || ab.CounterCostN > 0 || ab.CounterCostVariable || len(ab.Restrictions) > 0 {
			continue
		}
		cost := ab.ManaCost
		if ab.ChargedManaCost != nil {
			cost = *ab.ChargedManaCost
		}
		src := manaSource{units: abilityUnits(ab), input: manaValue(cost, 0)}
		src.in = manaCost{generic: src.input}
		if ab.LifeCost > 0 || strings.Contains(strings.ToLower(ab.Label), "damage to you") {
			for k := range src.units {
				src.units[k] |= manaPain
			}
		}
		if filterLands && src.input > 0 && colouredOnly(src.units) && (!haveFilter || src.net() > filter.net()) {
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
	return best, found && best.net() > 0
}

// colouredOnly reports whether every unit is a painless coloured mana.
func colouredOnly(units []uint8) bool {
	for _, u := range units {
		if u&manaAnyColor == 0 || u&manaPain != 0 {
			return false
		}
	}
	return len(units) > 0
}

// colouredAny reports whether any unit can be coloured mana.
func colouredAny(units []uint8) bool {
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
func abilityUnits(ab *protocol.ManaAbilityView) []uint8 {
	slot := 0
	optionMask := func() uint8 {
		if slot >= len(ab.ColorOptions) {
			return 0
		}
		var m uint8
		for _, l := range ab.ColorOptions[slot] {
			m |= colorMask(l)
		}
		slot++
		return m
	}
	if ab.Produced == "" {
		if m := optionMask(); m != 0 {
			return []uint8{m}
		}
		return []uint8{manaC}
	}
	var units []uint8
	for _, sym := range manaSymbols(ab.Produced) {
		n := 1
		var mask uint8
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
		return []uint8{manaC}
	}
	return units
}

// castAddsMana is the mana a card adds THIS turn once it resolves (§2):
// a noncreature permanent's repeatable mana ability (a rock enters
// untapped and has no summoning sickness, CR 302.6), a hasty creature's,
// and one mana of any colour for each land its purpose says enters
// untapped (PurposeView.lands_untapped, owner answer 4).
func castAddsMana(c *protocol.CardView, ps purposeSet, filterLands bool) []manaSource {
	var out []manaSource
	if !isLand(c) && isPermanentSpell(c) && (!isCreature(c) || hasKeyword(c, "haste")) {
		if src, ok := manaSourceOf(c, filterLands); ok && len(c.ManaAbilities) > 0 {
			out = append(out, src)
		}
	}
	for k := 0; k < ps.landsUntapped; k++ {
		out = append(out, manaSource{units: []uint8{manaAnyColor}})
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
func payMana(units []uint8, cost manaCost, later []uint8) ([]uint8, bool) {
	if cost.total() > len(units) {
		return units, false
	}
	colored := append([]uint8(nil), cost.colored...)
	// The most constrained symbol first: fewest units can pay it.
	sort.SliceStable(colored, func(i, j int) bool {
		return payers(units, colored[i]) < payers(units, colored[j])
	})
	for _, sym := range colored {
		pick := -1
		for k, u := range units {
			if u&sym == 0 {
				continue
			}
			if pick < 0 || spendBefore(u, units[pick], later) {
				pick = k
			}
		}
		if pick < 0 {
			return units, false
		}
		units = removeUnit(units, pick)
	}
	if cost.generic > len(units) {
		return units, false
	}
	engineOrder(units, later)
	return units[cost.generic:], true
}

// engineOrder sorts units into the order the auto-tapper spends them on
// generic mana, worst case first within its ties (payMana).
func engineOrder(units, later []uint8) {
	sort.SliceStable(units, func(i, j int) bool { return spendBefore(units[i], units[j], later) })
}

// spendBefore reports whether the auto-tapper spends a before b:
// painless first, then the fewest colours (colourless is none), and
// within a tie the one the later symbols need more.
func spendBefore(a, b uint8, later []uint8) bool {
	if pa, pb := a&manaPain != 0, b&manaPain != 0; pa != pb {
		return !pa
	}
	if ca, cb := bits.OnesCount8(a&manaAnyColor), bits.OnesCount8(b&manaAnyColor); ca != cb {
		return ca < cb
	}
	return demand(a, later) > demand(b, later)
}

// demand is how many of the later symbols u could pay.
func demand(u uint8, later []uint8) int {
	n := 0
	for _, sym := range later {
		if u&sym != 0 {
			n++
		}
	}
	return n
}

// payers is how many units can pay sym.
func payers(units []uint8, sym uint8) int {
	n := 0
	for _, u := range units {
		if u&sym != 0 {
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

func removeUnit(units []uint8, k int) []uint8 {
	units[k] = units[len(units)-1]
	return units[:len(units)-1]
}

// addSource activates src: a filter consumes its input as generic mana
// is paid (engineOrder), and is not run when there is not enough to
// feed it. A filter whose input names colours (PlanFilterLands) pays it
// as a cast would (payMana), and makes its fallback when it cannot.
func addSource(units []uint8, src manaSource, later []uint8) []uint8 {
	if len(src.in.colored) > 0 {
		rest, ok := payMana(append([]uint8(nil), units...), src.in, later)
		if !ok {
			return append(units, src.fallback...)
		}
		return append(rest, src.units...)
	}
	if src.input > 0 {
		if len(units) < src.input {
			return units
		}
		engineOrder(units, later)
		units = units[src.input:]
	}
	return append(units, src.units...)
}

// planMembers is the plan as the trace records it (§7).
func planMembers(moves []legal.Move, pl turnPlan) []aiseat.PlanMember {
	out := make([]aiseat.PlanMember, 0, len(pl.members))
	for _, c := range pl.members {
		out = append(out, aiseat.PlanMember{Index: c.index, Label: moves[c.index].Label})
	}
	return out
}

// planReason is the decision's reason (§7): "plan: Arcane Signet →
// Ornithopter of Paradise (+2.76)".
func planReason(pl turnPlan) string {
	names := make([]string, 0, len(pl.members))
	for _, c := range pl.members {
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
	return aiseat.Decision{Index: pl.members[0].index, Reason: planReason(pl)}, planMembers(moves, pl), pl, true, true
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
