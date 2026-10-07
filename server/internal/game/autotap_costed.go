package game

import (
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// autotap_costed.go is #2455: the auto-tapper funds a mana ability
// that owes MANA — a Signet's "{1}, {T}: Add {W}{B}", a filter land's
// "{W/U}, {T}" — from the plan's other sources.
//
// Before this, the planner left every such source out
// (autoTapAbilityAccepts' mana-cost bullet). The legal-move list
// therefore had no cast that needed a Signet's colours, smart autopass
// found nothing to play and passed the player's own main phase, and
// the card sat dimmed in hand although tapping a Plains for the {1}
// and the Signet for {W}{B} paid for it (#2455, #2456).
//
// The rules: CR 605.3a lets a player activate a mana ability while
// paying for a spell or ability, which is what every auto-tapped
// payment is, and nothing stops one of those activations being a
// Signet whose own {1} the mana of an earlier activation pays. What
// the planner has to respect is ORDER: a Signet's cost is paid when it
// is activated, out of mana already made. Its own mana cannot pay for
// it, and two Signets cannot pay for each other.
//
// # The model
//
// A costed plan is the ordinary plan with an ordered chain of costed
// sources K = k0, k1, …, k(m-1) behind it. Every source has a
// POSITION: mana floating before the plan began is -2, an ordinary
// source -1, chain member kj is j, and the cost being paid is m. Each
// mana requirement has an OWNER — kj's cost is owned by j, the spell's
// by m — and a slot may pay a requirement only when its source's
// position is below the owner's. That one inequality is the whole of
// the cycle rule: kj's mana pays only k(j+1) onward and the spell,
// never kj itself and never anything before it, so no chain can loop,
// and a Signet funding another Signet is fine exactly when the mana
// nets out.
//
// The coloured requirements (the spell's, and a filter land's {W/U})
// are placed slot by slot by depth-first backtracking, as the ordinary
// solver places them, with the eligibility test added. The generic
// ones are then counted owner by owner, lowest first: the slots
// eligible for owner j are a subset of those eligible for every later
// owner, so taking any eligible free slot is as good as any other, and
// an unused ordinary source is recruited (position -1, eligible for
// all) when a prefix runs short.
//
// The chains are tried shortest first (one Signet before two), and
// among chains of one length in the order sortTapSources ranks the
// candidates, so a plan never taps a Signet it does not need and
// reaches for the same tiers the ordinary planner does. Two candidates
// with the same shape (two Orzhov Signets) are tried in one order
// only. Every node shares AutoTapBudget, and a search that exhausts it
// answers "no plan", the same answer for the move list and the
// payment, because both ask this function.
//
// # The plan and the executor
//
// The plan is EXPLICIT where the ordinary one is not: every entry
// carries the colour it books for each mana it makes (plannedTap.Slots)
// and a costed entry the tokens that fund it (plannedTap.FundedBy, by
// source and colour). The executor (materializePlanLocked) runs the
// ordinary entries first, then the chain in order, and pays each
// costed entry's cost out of exactly the tokens the plan named, before
// tapping it — the component order ActivateManaAbility pays in (mana,
// then the tap). The two halves therefore cannot disagree about which
// mana paid the Signet, which is what would otherwise spend the {W}
// the spell needed on the Signet's {1}.
//
// planFundsLocked still runs every auto-tapped payment on a clone
// first, so a plan the executor could not carry out is refused with
// nothing tapped.
//
// # What it does not plan
//
// Weaker than printed, never stronger — each one is a card the player
// still activates by hand from its ability menu:
//
//   - an ability whose output is computed (ProducedFunc, a derived or
//     paid-for output): Cabal Coffers, Nykthos, Doubling Cube, which
//     reads the very pool the plan is filling;
//   - an ability with spend riders or a pre-rider (Opal Palace);
//   - a cost with {X}, Phyrexian, snow or a widened symbol, or one
//     that makes fewer mana than it costs;
//   - any board with a CR 614 mana replacement on it (Mana
//     Reflection): the executor's slots would not be the planner's;
//   - restricted or rider-carrying floating mana, which the plan
//     neither counts nor spends;
//   - a chain longer than maxCostedChain.
//
// And it runs only when the ordinary planner has found no plan, so
// every payment that planner already made is made exactly as before.

// maxCostedChain caps how many costed sources one plan chains.
const maxCostedChain = 4

// autoTapCostedCost reports whether a battlefield mana ability owes
// mana a costed plan can fund (#2455), and returns that mana priced
// through the CR 601.2f pass. Every exclusion autoTapAbilityAccepts
// makes besides the mana cost is made here too.
//
// Caller must hold g.mu.
func (g *Game) autoTapCostedCost(asker uuid.UUID, source Card, a ManaAbilityShape) (ParsedCost, bool) {
	if a.ManaCost == "" || !g.autoTapAcceptsBesidesManaCost(asker, source, a) {
		return ParsedCost{}, false
	}
	// The planner books a static output. A computed one may read what
	// the plan changes (Doubling Cube reads the pool), and a pre-rider
	// or a spend rider is a closure it cannot price.
	if a.ProducedFunc != nil || a.DerivedMatch != nil || a.ProducedForPaid != nil ||
		a.PreRider != nil || len(a.SpendRiders) > 0 {
		return ParsedCost{}, false
	}
	priced, err := g.ManaAbilityManaCostForEffect(asker, source, a)
	if err != nil || priced.Empty() {
		// Priced to nothing is the ordinary planner's source (#1191).
		return ParsedCost{}, false
	}
	if priced.XSlots > 0 || priced.HasPhyrexian || priced.HasSnow {
		return ParsedCost{}, false
	}
	for _, r := range priced.Required {
		if r.Phyrexian || r.HasNumericAlt || r.Snow || r.AnyMana || len(r.Options) == 0 {
			return ParsedCost{}, false
		}
	}
	return priced, true
}

// autoTapCostedAbilitiesFor is autoTapAbilitiesFor for the abilities
// that owe mana (#2455): every one autoTapCostedCost accepts, with its
// ref and its priced cost.
//
// Caller must hold g.mu.
func (g *Game) autoTapCostedAbilitiesFor(asker uuid.UUID, source Card) []autoTapCandidate {
	abilities, origins := ManaAbilitiesWithOrigins(source)
	var out []autoTapCandidate
	for i := range abilities {
		cost, ok := g.autoTapCostedCost(asker, source, abilities[i])
		if !ok {
			continue
		}
		o := origins.At(i)
		out = append(out, autoTapCandidate{ab: abilities[i], ref: o.Ref, granted: o.Granted(), cost: cost})
	}
	return out
}

// plannedAbilityLocked is the ability a plan entry booked, re-found
// through the picker that offered it: the costed picker for a Costed
// entry (#2455), autoTapAbilityForRef for every other. Nil when the
// ability is no longer acceptable.
//
// Caller must hold g.mu.
func (g *Game) plannedAbilityLocked(asker uuid.UUID, source Card, planned plannedTap) *ManaAbilityShape {
	if !planned.Costed {
		return g.autoTapAbilityForRef(asker, source, planned.Ref)
	}
	for _, cand := range g.autoTapCostedAbilitiesFor(asker, source) {
		if planned.Ref == "" || cand.ref == planned.Ref {
			ab := cand.ab
			return &ab
		}
	}
	return nil
}

// Positions of a costed solver's sources (see the file comment).
const (
	posFloating = -2
	posOrdinary = -1
)

// costedSrc is one source of a costed search: a candidate and where it
// stands in the activation order.
type costedSrc struct {
	tapSource
	pos int
}

// ownedReq is one coloured requirement and the position that owes it.
type ownedReq struct {
	req   ColorRequirement
	owner int
}

// costedSolver is the state of one chain's search.
type costedSolver struct {
	srcs     []costedSrc
	used     []bool
	consumed [][]bool
	owner    [][]int
	color    [][]string
	taken    map[uuid.UUID]bool
	reqs     []ownedReq
	m        int
	// spellGeneric is the cost's own generic, X included.
	spellGeneric int
	// costGeneric[j] is chain member j's generic.
	costGeneric []int
	deferred    int
	pain        int
	budget      *int
}

// autoTapCostedLocked is #2455's planner: a plan for `cost` that
// funds one or more costed sources from the plan's other sources and
// the floating pool. Called by autoTapTopUpLocked only when the
// ordinary planner found no plan for any shortfall. Read-only.
//
// The plan pays the WHOLE cost, not a shortfall: the floating mana is
// a source of the plan (position -2), because it may be what pays a
// Signet's {1}.
//
// Caller must hold g.mu.
func (g *Game) autoTapCostedLocked(
	p *Player,
	cost ParsedCost,
	xValue int,
	excluded map[uuid.UUID]bool,
	prefer ManaSourceKinds,
) (tapPlan, bool) {
	if p == nil || g.producesManaReplacementsExistLocked() {
		return nil, false
	}
	costed := gatherSources(g, p.ID, excluded, prefer, true)
	costed = netPositiveCosted(costed)
	if len(costed) == 0 {
		return nil, false
	}
	sortTapSources(costed)
	ordinary := gatherTapSources(g, p.ID, excluded, prefer)
	sortTapSources(ordinary)
	floating := floatingSources(p.ManaPool)

	budget := AutoTapBudget
	painBudget := painBudgetFor(g, p.ID)
	keys := make([]string, len(costed))
	for i := range costed {
		keys[i] = costedShapeKey(costed[i])
	}
	chosen := make([]bool, len(costed))
	var seq []int

	var try func(size int) (tapPlan, bool)
	try = func(size int) (tapPlan, bool) {
		if budget <= 0 {
			return nil, false
		}
		if len(seq) == size {
			return solveCostedChain(costed, seq, ordinary, floating, cost, xValue, painBudget, &budget)
		}
		for i := range costed {
			if chosen[i] || chainHasCard(costed, seq, costed[i].CardID) {
				continue
			}
			// Two candidates of one shape are one choice: take them in
			// index order only.
			if earlierSameShapeFree(keys, chosen, i) {
				continue
			}
			chosen[i] = true
			seq = append(seq, i)
			if plan, ok := try(size); ok {
				return plan, true
			}
			seq = seq[:len(seq)-1]
			chosen[i] = false
			if budget <= 0 {
				return nil, false
			}
		}
		return nil, false
	}
	for size := 1; size <= len(costed) && size <= maxCostedChain; size++ {
		if plan, ok := try(size); ok {
			return plan, true
		}
		if budget <= 0 {
			break
		}
	}
	return nil, false
}

// netPositiveCosted drops a costed candidate that makes fewer mana
// than it costs: activating it can only shrink the pool, so no plan
// needs it.
func netPositiveCosted(in []tapSource) []tapSource {
	out := in[:0]
	for _, s := range in {
		if len(s.Slots) >= s.Cost.Generic+len(s.Cost.Required) {
			out = append(out, s)
		}
	}
	return out
}

// floatingSources is the floating pool as costed-solver sources: one
// single-slot source per unrestricted, rider-free token. Restricted
// mana is left out of the plan's arithmetic entirely; it stays in the
// pool for the payment that may spend it.
func floatingSources(pool ManaPool) []tapSource {
	var out []tapSource
	for _, tok := range pool {
		if len(tok.Restrictions) > 0 || len(tok.Riders) > 0 || tok.Color == "" {
			continue
		}
		out = append(out, tapSource{
			CardID: tok.Source,
			Slots:  []ProducedManaEntry{{Options: []string{tok.Color}}},
		})
	}
	return out
}

// costedShapeKey names everything about a costed candidate the search
// can tell apart, so two Orzhov Signets are tried in one order only.
func costedShapeKey(s tapSource) string {
	var b strings.Builder
	b.WriteString(s.Cost.String())
	for _, slot := range s.Slots {
		b.WriteString("|")
		b.WriteString(strings.Join(slot.Options, ","))
	}
	b.WriteString("|" + s.OneColor + "|" + strconv.Itoa(s.Pain))
	for _, bit := range []bool{s.Frozen, s.Sacrifices, s.SacrificesCreature, s.Wanted, s.GrantedCreature} {
		if bit {
			b.WriteString("1")
		} else {
			b.WriteString("0")
		}
	}
	return b.String()
}

// earlierSameShapeFree reports whether a candidate before i has i's
// shape and is not in the chain yet.
func earlierSameShapeFree(keys []string, chosen []bool, i int) bool {
	for j := 0; j < i; j++ {
		if keys[j] == keys[i] && !chosen[j] {
			return true
		}
	}
	return false
}

// chainHasCard reports whether a permanent is already in the chain
// (one permanent offers one candidate per ability, and taps once).
func chainHasCard(costed []tapSource, seq []int, id uuid.UUID) bool {
	for _, i := range seq {
		if costed[i].CardID == id {
			return true
		}
	}
	return false
}

// solveCostedChain searches for a plan with this chain of costed
// sources behind the ordinary ones.
func solveCostedChain(
	costed []tapSource,
	seq []int,
	ordinary, floating []tapSource,
	cost ParsedCost,
	xValue, painBudget int,
	budget *int,
) (tapPlan, bool) {
	m := len(seq)
	s := &costedSolver{
		m:            m,
		taken:        map[uuid.UUID]bool{},
		spellGeneric: cost.Generic + cost.XSlots*xValue,
		costGeneric:  make([]int, m),
		pain:         painBudget,
		budget:       budget,
	}
	for _, i := range seq {
		s.taken[costed[i].CardID] = true
	}
	for _, f := range floating {
		s.add(costedSrc{tapSource: f, pos: posFloating}, true)
	}
	for _, o := range ordinary {
		if s.taken[o.CardID] {
			continue
		}
		s.add(costedSrc{tapSource: o, pos: posOrdinary}, false)
	}
	for j, i := range seq {
		s.add(costedSrc{tapSource: costed[i], pos: j}, true)
		s.pain -= costed[i].Pain
		s.costGeneric[j] = costed[i].Cost.Generic
		for _, r := range costed[i].Cost.Required {
			s.reqs = append(s.reqs, ownedReq{req: r, owner: j})
		}
	}
	if s.pain < 0 {
		return nil, false
	}
	for _, r := range widenedLast(cost.Required) {
		s.reqs = append(s.reqs, ownedReq{req: r, owner: m})
	}
	if !s.colored(0) {
		return nil, false
	}
	return s.plan(), true
}

// add appends a source; `used` marks one the plan spends whatever the
// search decides (floating mana, a chain member).
func (s *costedSolver) add(src costedSrc, used bool) {
	s.srcs = append(s.srcs, src)
	s.used = append(s.used, used)
	n := len(src.Slots)
	s.consumed = append(s.consumed, make([]bool, n))
	owners := make([]int, n)
	for k := range owners {
		owners[k] = -1
	}
	s.owner = append(s.owner, owners)
	s.color = append(s.color, make([]string, n))
}

func (s *costedSolver) use(i int) {
	s.used[i] = true
	s.taken[s.srcs[i].CardID] = true
	s.pain -= s.srcs[i].Pain
}

func (s *costedSolver) unuse(i int) {
	s.used[i] = false
	delete(s.taken, s.srcs[i].CardID)
	s.pain += s.srcs[i].Pain
}

func (s *costedSolver) book(i, k, owner int, color string) {
	s.consumed[i][k] = true
	s.owner[i][k] = owner
	s.color[i][k] = color
}

func (s *costedSolver) unbook(i, k int) {
	s.consumed[i][k] = false
	s.owner[i][k] = -1
	s.color[i][k] = ""
}

// usable reports whether an unused ordinary source may join the plan:
// its permanent is not already in it, and the pain budget covers it.
func (s *costedSolver) usable(i int) bool {
	src := s.srcs[i]
	return src.pos == posOrdinary && !s.taken[src.CardID] && src.Pain <= s.pain
}

// colored places coloured requirement ri and every one after it, then
// the generic ones (see the file comment).
func (s *costedSolver) colored(ri int) bool {
	if *s.budget <= 0 {
		return false
	}
	*s.budget--
	if ri == len(s.reqs) {
		return s.generic()
	}
	r := s.reqs[ri]
	for i := range s.srcs {
		// The cycle rule: only mana made before the owner pays it.
		if s.srcs[i].pos >= r.owner {
			continue
		}
		wasUsed := s.used[i]
		if !wasUsed && !s.usable(i) {
			continue
		}
		for _, k := range matchingFreeSlots(s.srcs[i].tapSource, s.consumed[i], r.req) {
			if !wasUsed {
				s.use(i)
			}
			s.book(i, k, r.owner, slotColorFor(s.srcs[i].Slots[k], r.req))
			if s.colored(ri + 1) {
				return true
			}
			s.unbook(i, k)
			if !wasUsed {
				s.unuse(i)
			}
			if *s.budget <= 0 {
				return false
			}
		}
	}
	// #1600: a widened symbol of the cost being paid that no printed
	// colour can take is paid as generic, as the ordinary solver does.
	if r.req.AnyMana && r.owner == s.m {
		s.deferred++
		if s.colored(ri + 1) {
			return true
		}
		s.deferred--
	}
	return false
}

// generic counts the generic requirements onto free slots, owner by
// owner, recruiting ordinary sources when a prefix runs short. On
// failure it undoes everything it did, so the coloured search can try
// its next placement.
func (s *costedSolver) generic() bool {
	type booking struct{ i, k int }
	var booked []booking
	var recruited []int
	undo := func() {
		for n := len(booked) - 1; n >= 0; n-- {
			s.unbook(booked[n].i, booked[n].k)
		}
		for n := len(recruited) - 1; n >= 0; n-- {
			s.unuse(recruited[n])
		}
	}
	plain := make([]tapSource, len(s.srcs))
	for i := range s.srcs {
		plain[i] = s.srcs[i].tapSource
	}
	order := orderUnusedByGenericPreference(plain, s.used)
	next := 0
	for owner := 0; owner <= s.m; owner++ {
		units := s.spellGeneric + s.deferred
		if owner < s.m {
			units = s.costGeneric[owner]
		}
		for units > 0 {
			i, k := s.freeSlotFor(owner)
			if i < 0 {
				// Recruit the next ordinary source the generic order
				// prefers; position -1 is eligible for every owner.
				r := -1
				for ; next < len(order); next++ {
					if c := order[next]; !s.used[c] && s.usable(c) {
						r = c
						next++
						break
					}
				}
				if r < 0 {
					undo()
					return false
				}
				s.use(r)
				recruited = append(recruited, r)
				continue
			}
			s.book(i, k, owner, genericColorFor(s.srcs[i].Slots[k]))
			booked = append(booked, booking{i, k})
			units--
		}
	}
	return true
}

// freeSlotFor is a free slot of a used source that may pay a generic
// mana owner `owner` owes, colourless first, or (-1, -1).
func (s *costedSolver) freeSlotFor(owner int) (int, int) {
	bi, bk := -1, -1
	for i := range s.srcs {
		if !s.used[i] || s.srcs[i].pos >= owner {
			continue
		}
		for k, slot := range s.srcs[i].Slots {
			if s.consumed[i][k] {
				continue
			}
			if len(slot.Options) == 1 && slot.Options[0] == "C" {
				return i, k
			}
			if bi < 0 {
				bi, bk = i, k
			}
		}
	}
	return bi, bk
}

// plan turns a solved search into the explicit plan the executor
// carries out: the ordinary sources, then the chain in order.
func (s *costedSolver) plan() tapPlan {
	entry := func(i int) plannedTap {
		src := s.srcs[i]
		e := plannedTap{CardID: src.CardID, OneColor: src.OneColor, Ref: src.Ref}
		e.Slots = make([]plannedSlot, len(src.Slots))
		for k, slot := range src.Slots {
			c := s.color[i][k]
			if c == "" {
				c = genericColorFor(slot)
			}
			o := s.owner[i][k]
			e.Slots[k] = plannedSlot{Color: c, Funds: o >= 0 && o < s.m}
		}
		return e
	}
	var out tapPlan
	for i, src := range s.srcs {
		if src.pos == posOrdinary && s.used[i] {
			out = append(out, entry(i))
		}
	}
	for j := 0; j < s.m; j++ {
		for i, src := range s.srcs {
			if src.pos != j {
				continue
			}
			e := entry(i)
			e.Costed = true
			for fi := range s.srcs {
				for k, o := range s.owner[fi] {
					if o == j {
						e.FundedBy = append(e.FundedBy, fundToken{Source: s.srcs[fi].CardID, Color: s.color[fi][k]})
					}
				}
			}
			out = append(out, e)
		}
	}
	return out
}

// slotColorFor is the colour a slot makes to pay a requirement: its
// first option the requirement admits by a printed colour.
func slotColorFor(slot ProducedManaEntry, req ColorRequirement) string {
	for _, opt := range slot.Options {
		if matchColor(opt, req.Options) {
			return opt
		}
	}
	return slot.Options[0]
}

// genericColorFor is the colour a slot makes for generic mana or for
// nothing in particular: colourless when it can, else its first
// option, which is the identity-first order the activation offers.
func genericColorFor(slot ProducedManaEntry) string {
	for _, opt := range slot.Options {
		if opt == "C" {
			return opt
		}
	}
	return slot.Options[0]
}

// plannedSlotColor is the colour an explicit plan (#2455) booked for
// slot si of an entry, when the plan has one for it and the slot still
// offers it. A plan whose slot count no longer matches the source's
// (a one-colour pick, or an ordinary plan with no Slots at all) has
// none, and the executor colour-picks greedily as before.
func plannedSlotColor(planned plannedTap, si, n int, options []string) (plannedSlot, bool) {
	if len(planned.Slots) != n || si < 0 || si >= n {
		return plannedSlot{}, false
	}
	ps := planned.Slots[si]
	if !colorOffered(options, ps.Color) {
		return plannedSlot{}, false
	}
	return ps, true
}

// payPlannedManaCostLocked pays a Costed plan entry's mana cost (#2455)
// out of `p`'s pool, the way ActivateManaAbility pays it: priced
// through the CR 601.2f pass, widened by a spend grant, and spent under
// the activation's own context. It spends the tokens the plan named
// (`funded`) when they pay the cost, and the pool as a whole only when
// they do not, so the mana the plan meant for the spell stays in the
// pool. Returns the tokens spent, or false with the pool untouched.
//
// Caller must hold g.mu.
func (g *Game) payPlannedManaCostLocked(p *Player, source Card, ab *ManaAbilityShape, funded []fundToken) ([]ManaToken, bool) {
	priced, err := g.ManaAbilityManaCostForEffect(p.ID, source, *ab)
	if err != nil {
		return nil, false
	}
	ctx := ManaSpendForAbility(source)
	cost := g.costAsPaidByLocked(p.ID, ctx, priced, 0)
	if cost.Empty() {
		return nil, true
	}
	rest := append(ManaPool(nil), p.ManaPool...)
	var named ManaPool
	for _, f := range funded {
		for i, tok := range rest {
			if tok.Source != f.Source || tok.Color != f.Color || len(tok.Restrictions) > 0 || len(tok.Riders) > 0 {
				continue
			}
			named = append(named, tok)
			rest = append(rest[:i], rest[i+1:]...)
			break
		}
	}
	if spent, ok := named.SpendManaFor(cost, 0, ctx); ok {
		pool := append(rest, named...)
		if len(pool) == 0 {
			pool = nil
		}
		p.ManaPool = pool
		return spent, true
	}
	return p.ManaPool.SpendManaFor(cost, 0, ctx)
}
