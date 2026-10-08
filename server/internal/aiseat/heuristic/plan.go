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
//     cheapest first, then the rest, highest value first. The bot makes
//     the first move; the next window plans again. Nothing is stored.
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
// per mana, and the generic mana it consumes first (a filter, a
// Signet's {1}).
type manaSource struct {
	units []uint8
	input int
}

func (s manaSource) net() int { return len(s.units) - s.input }

// planCandidate is one card's best plan-eligible cast in this window.
type planCandidate struct {
	index int
	card  *protocol.CardView
	value float64
	// premium is the ramp premium valueOf included, and amount the mana
	// a turn the card adds, which §3 shares across the set.
	premium float64
	amount  int
	cost    manaCost
	// adds is the mana the card makes this turn once it resolves.
	adds []manaSource
	// class is §4's order: 0 adds mana, 1 draws or tutors, 2 the rest.
	class int
	// sweep is a declared sweep (ADR 0126 §4). Its price is what it
	// removes from the board as it stands, so among the rest it goes
	// before any permanent the plan adds.
	sweep bool
}

// turnPlan is the plan chosen in one window.
type turnPlan struct {
	members []*planCandidate
	value   float64
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
// members is feasible.
func (p *Policy) planTurn(ctx context.Context, st *state, moves []legal.Move, vals []float64) (turnPlan, bool) {
	cands := p.planCandidates(st, moves, vals)
	if len(cands) < 2 {
		return turnPlan{}, false
	}
	base := st.manaAvailable()
	maxMana := len(base)
	for _, c := range cands {
		for _, a := range c.adds {
			if n := a.net(); n > 0 {
				maxMana += n
			}
		}
	}
	want := st.rampWants()
	sources := st.manaSourcesTotal()

	n := len(cands)
	bestMask, bestVal := uint(0), 0.0
	units := make([]uint8, 0, maxMana)
	for mask := uint(1); mask < 1<<n; mask++ {
		if mask%64 == 0 && ctx.Err() != nil {
			break
		}
		size := bits.OnesCount(mask)
		if size < 2 {
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
			shared := p.cfg.RampPerMana * float64(min(amount, p.planDeficit(want, sources, cands, mask)))
			value += min(premium, shared)
		}
		// Ties go to the smaller set (§3), then to the set found first.
		better := bestMask == 0 || value > bestVal+planEpsilon ||
			(value > bestVal-planEpsilon && size < bits.OnesCount(bestMask))
		if !better {
			continue
		}
		if !planFeasible(base, cands, mask, units) {
			continue
		}
		bestMask, bestVal = mask, value
	}
	if bestMask == 0 {
		return turnPlan{}, false
	}
	out := turnPlan{value: bestVal}
	for i := 0; i < n; i++ {
		if bestMask&(1<<i) != 0 {
			out.members = append(out.members, cands[i])
		}
	}
	return out, true
}

// planCandidates is each card's best plan-eligible cast, in §4's order,
// the PlanMaxCards most valuable of them.
func (p *Policy) planCandidates(st *state, moves []legal.Move, vals []float64) []*planCandidate {
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
		c := &planCandidate{index: i, card: card, value: vals[i], cost: cost}
		if byCard[card.InstanceID] == nil {
			order = append(order, card.InstanceID)
		}
		byCard[card.InstanceID] = c
	}
	if len(order) < 2 {
		return nil
	}
	out := make([]*planCandidate, 0, len(order))
	for _, id := range order {
		out = append(out, byCard[id])
	}
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
		m := moves[c.index]
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
		c.adds = castAddsMana(c.card, ps)
		switch {
		case len(c.adds) > 0:
			c.class = 0
		case ps.draws > 0 || ps.tutors > 0:
			c.class = 1
		default:
			c.class = 2
		}
		c.sweep = p.cfg.PriceSweeps && len(ps.sweeps) > 0
	}
	// §4's order, so a set's members are always taken in it.
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.class != b.class {
			return a.class < b.class
		}
		if a.class < 2 && a.cost.total() != b.cost.total() {
			return a.cost.total() < b.cost.total()
		}
		if a.sweep != b.sweep {
			return a.sweep
		}
		if a.value != b.value {
			return a.value > b.value
		}
		return a.index < b.index
	})
	return out
}

// planFeasible reports whether the members in mask can be paid in §4's
// order from base, each member's mana arriving for the ones after it.
// scratch is reused between calls.
func planFeasible(base []uint8, cands []*planCandidate, mask uint, scratch []uint8) bool {
	units := append(scratch[:0], base...)
	for i, c := range cands {
		if mask&(1<<i) == 0 {
			continue
		}
		// The coloured symbols of the members still to come: generic
		// mana is paid around them.
		var later []uint8
		for j := i + 1; j < len(cands); j++ {
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
// can activate, filters already run.
func (st *state) manaAvailable() []uint8 {
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
		src, ok := manaSourceOf(c)
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
// colour, and a filter row loses to a plain one that nets as much. A
// row whose mana may be spent only on some spells (Delighted Halfling's
// colours, Ancient Ziggurat) is left out: the model cannot tell which
// casts it pays for. A
// land with no mana rows (a fetchland) makes no mana: every land that
// taps for mana carries a row, its intrinsic ones included.
func manaSourceOf(c *protocol.CardView) (manaSource, bool) {
	var best manaSource
	found := false
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
		switch {
		case !found || src.net() > best.net() || (src.net() == best.net() && src.input < best.input):
			best, found = src, true
		case src.net() == best.net() && src.input == best.input && len(src.units) == len(best.units):
			for k := range best.units {
				best.units[k] |= src.units[k]
			}
		}
	}
	return best, found && best.net() > 0
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
func castAddsMana(c *protocol.CardView, ps purposeSet) []manaSource {
	var out []manaSource
	if !isLand(c) && isPermanentSpell(c) && (!isCreature(c) || hasKeyword(c, "haste")) {
		if src, ok := manaSourceOf(c); ok && len(c.ManaAbilities) > 0 {
			out = append(out, src)
		}
	}
	for k := 0; k < ps.landsUntapped; k++ {
		out = append(out, manaSource{units: []uint8{manaAnyColor}})
	}
	return out
}

// payMana pays cost out of units (§2): coloured symbols first, each
// from the most constrained unit that can pay it, then generic from the
// units the later symbols need least (spareFirst). It reorders units in
// place.
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
			if pick < 0 || bits.OnesCount8(u) < bits.OnesCount8(units[pick]) {
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
	spareFirst(units, later)
	return units[cost.generic:], true
}

// spareFirst orders units so the ones to spend on generic mana come
// first: those the later coloured symbols could use least, then the
// least flexible.
func spareFirst(units, later []uint8) {
	demand := func(u uint8) int {
		n := 0
		for _, sym := range later {
			if u&sym != 0 {
				n++
			}
		}
		return n
	}
	sort.SliceStable(units, func(i, j int) bool {
		di, dj := demand(units[i]), demand(units[j])
		if di != dj {
			return di < dj
		}
		return bits.OnesCount8(units[i]) < bits.OnesCount8(units[j])
	})
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
// is paid (spareFirst), and is not run when there is not enough to feed
// it.
func addSource(units []uint8, src manaSource, later []uint8) []uint8 {
	if src.input > 0 {
		if len(units) < src.input {
			return units
		}
		spareFirst(units, later)
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
func (p *Policy) decidePlan(ctx context.Context, st *state, moves []legal.Move, vals []float64,
	best int, bestVal, threshold float64, leftover bool) (aiseat.Decision, []aiseat.PlanMember, bool) {
	if !p.cfg.PlanTurnMana || !st.sorcerySpeed || best < 0 || moves[best].Kind == legal.KindLand || ctx.Err() != nil {
		return aiseat.Decision{}, nil, false
	}
	pl, ok := p.planTurn(ctx, st, moves, vals)
	if !ok || len(pl.members) < 2 || pl.value <= bestVal+planEpsilon {
		return aiseat.Decision{}, nil, false
	}
	bar := threshold
	if leftover {
		all := true
		for _, c := range pl.members {
			if !p.leftoverEligible(st, moves[c.index]) {
				all = false
				break
			}
		}
		if all {
			bar = p.cfg.LeftoverThreshold
		}
	}
	if pl.value <= bar {
		return aiseat.Decision{}, nil, false
	}
	return aiseat.Decision{Index: pl.members[0].index, Reason: planReason(pl)}, planMembers(moves, pl), true
}
