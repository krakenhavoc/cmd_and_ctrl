package game

import (
	"sort"
	"strings"

	"github.com/google/uuid"
)

// autotap_wish.go — #3853, ADR 0068 amendment 2026-10-15: auto-pay picks
// the mana that DOES something.
//
// #1212 gave the auto-tapper one wish: the spell being cast could name
// the KINDS of source it reads back (Hired Hexblade → a Treasure). This
// file widens the same wish in the two directions that left the choice of
// mana to chance:
//
//   - SOURCE-SIDE RIDERS. Path of Ancestry scries only when its mana pays
//     a creature spell that shares a type with the commander; Hall of the
//     Bandit Lord hastes only a creature; Boseiju protects only an
//     instant or sorcery. A source whose rider would fire on THIS spell is
//     preferred, and one whose rider would be wasted on it is held back
//     behind ordinary sources (tapSource.RiderRank).
//   - COLOUR READERS. A spell that counts the colours that paid for it —
//     converge and sunburst (Spec.WantsDistinctColors), "if {R} was
//     spent" (Spec.WantsColors) — prefers a plan whose mana is the colours
//     it reads (tapSource.ReadsColor, colorSpread).
//
// And the same preferences choose which FLOATING mana is spent
// (spendOrder, spendGenericWished), so a rider token is spent on the
// spell it benefits and kept from one it would be wasted on.
//
// # The bound is #1212's, unchanged
//
// ORDERING ONLY. Nothing here adds or drops a candidate source or a pool
// token, and nothing here is consulted by a payability test, so the
// legal-move enumerator (which asks the same planner for a bool) and the
// cast gate keep agreeing: a cast is offered exactly when it can be
// paid. Every key is a tiebreak below the planner's own thrift tiers —
// the sacrifice, hand, pain, energy and frozen tiers — because a rider or
// a colour reading is a bonus, and a bonus does not cost a Treasure.
//
// Consequently a source the planner REFUSES stays refused (a restricted
// mana ability — Cavern of Souls' coloured half, a Delighted Halfling's
// legendary-only half — is "a decision, not a resource", autotap.go's
// autoTapAbilityAccepts), however well its rider would suit the spell.
// The floating mana such a source made is ranked like any other token.
//
// The wish is derived from the ManaSpendContext every cast payment already
// threads through the planner and the pool, so no payer had to learn a new
// parameter: ManaSpendForCast snapshots the card's declarations into
// ManaSpendContext.wish.

// CatalogWantsColors is the catalog hook for the colours a spell's own
// text reads back from its payment ("if {R} was spent to cast it",
// Moonhold, Gruul Scrapper). Uppercase WUBRG letters; nil for a card that
// reads none. Nil hook means no catalog is wired.
//
// A declaration on the spec for the reason WantsDistinctColors and
// WantsManaFrom are: a scan of the oracle text quietly stops matching
// when a card words the clause differently.
var CatalogWantsColors func(oracleID string) []string

// spendWish is what a CAST's object reads back from its payment,
// snapshotted when the ManaSpendContext is built. Plain data — an ID, a
// bitset, two lists and a map of answers — so a context stored in game
// state (AssistPrompt.spend) carries no Game and no closure with it.
//
// Casts only. An activated ability's payment names no wish: a rider reads
// "when that mana is spent to cast", and the ability-cost paths were not
// part of #3853.
type spendWish struct {
	// paidID is the spell being cast. A spend rider's Condition is asked
	// of the card it names, found in whatever zone holds it when the
	// question is asked.
	paidID uuid.UUID

	// kinds, colors and distinct are the card's declarations
	// (WantsManaFrom, WantsColors, WantsDistinctColors).
	kinds    ManaSourceKinds
	colors   []string
	distinct bool

	// verdicts is the answer to each conditioned rider key on the
	// payer's floating mana (riderKey), stamped by
	// stampRiderVerdictsLocked where a payment is made. A Condition
	// reads the board, so it cannot be answered from the context alone;
	// a key with no entry is unanswered and its rider ranks neutral.
	verdicts map[string]bool
}

// newSpendWish snapshots what the spell `c` reads back.
func newSpendWish(c Card) *spendWish {
	w := &spendWish{paidID: c.InstanceID}
	key := catalogKeyOf(&c)
	if CatalogWantsManaFrom != nil {
		w.kinds = CatalogWantsManaFrom(key)
	}
	if CatalogWantsDistinctColors != nil {
		w.distinct = CatalogWantsDistinctColors(key)
	}
	if CatalogWantsColors != nil {
		w.colors = wantsColorList(CatalogWantsColors(key))
	}
	return w
}

// conditionedRider is the registry lookup of a rider's Condition: the key
// it is memoised under and the function, or ok=false for a rider that is
// not conditioned or names a key this binary does not register.
func conditionedRider(r ManaSpendRider) (key string, fn func(g *Game, controller uuid.UUID, paidFor Card) bool, registered bool) {
	switch {
	case r.Kind == ManaRiderTrigger:
		t, ok := ManaSpendTriggerFor(r.Trigger)
		if !ok {
			return "", nil, false
		}
		return "t:" + r.Trigger, t.Condition, true
	case r.Kind == ManaRiderEntersWithCounters && r.Count != "":
		c, ok := ManaRiderCountFor(r.Count)
		if !ok {
			return "", nil, false
		}
		return "c:" + r.Count, c.Condition, true
	}
	return "", nil, true
}

// conditionAsker answers a rider's Condition for the spell being cast:
// (result, true), or (_, false) when it cannot say.
type conditionAsker func(key string, fn func(g *Game, controller uuid.UUID, paidFor Card) bool) (bool, bool)

// verdictAsker reads the answers stamped on the context.
func (ctx ManaSpendContext) verdictAsker() conditionAsker {
	return func(key string, _ func(*Game, uuid.UUID, Card) bool) (bool, bool) {
		if ctx.wish == nil {
			return false, false
		}
		v, ok := ctx.wish.verdicts[key]
		return v, ok
	}
}

// stampRiderVerdictsLocked returns ctx with the Conditions of the
// riders on `controller`'s floating mana answered against this game, so
// the pool solver can tell mana whose rider fires on this spell from mana
// whose rider it would waste. A context with no wish, or a pool with no
// conditioned rider, is returned as it came.
//
// Caller must hold g.mu.
func (g *Game) stampRiderVerdictsLocked(ctx ManaSpendContext, controller uuid.UUID) ManaSpendContext {
	if ctx.wish == nil {
		return ctx
	}
	p := g.playerByIDLocked(controller)
	if p == nil {
		return ctx
	}
	var verdicts map[string]bool
	var card Card
	haveCard := false
	for _, tok := range p.ManaPool {
		for _, r := range tok.Riders {
			key, fn, ok := conditionedRider(r)
			if !ok || fn == nil {
				continue
			}
			if _, done := verdicts[key]; done {
				continue
			}
			if !haveCard {
				card, haveCard = g.LookupCardForEffect(ctx.wish.paidID)
				if !haveCard {
					return ctx
				}
			}
			if verdicts == nil {
				verdicts = map[string]bool{}
			}
			verdicts[key] = fn(g, controller, card)
		}
	}
	if verdicts == nil {
		return ctx
	}
	w := *ctx.wish
	w.verdicts = verdicts
	ctx.wish = &w
	return ctx
}

// riderFires reports whether rider `r` would fire if its mana paid for
// this context's spell — the SAME test applyManaSpendRidersLocked makes
// at the spend (the rider's When filter against the context, then the
// registered Condition) — and whether the answer is known. A rider the
// planner cannot judge (a mark that never fires, a key this binary does
// not know, a Condition nobody could answer) is unknown and ranks
// neutral: the weaker claim.
func (ctx ManaSpendContext) riderFires(r ManaSpendRider, ask conditionAsker) (fires, known bool) {
	if ctx.wish == nil {
		return false, false
	}
	switch r.Kind {
	case ManaRiderKeepUntilEndOfTurn, ManaRiderKeepUntilEndOfCombat:
		// A mark the step-boundary sweep reads; it never fires at a spend.
		return false, false
	}
	if !ctx.allows(r.When) {
		return false, true
	}
	key, fn, registered := conditionedRider(r)
	if !registered {
		return false, false
	}
	if fn == nil {
		return true, true
	}
	if ask == nil {
		return false, false
	}
	got, ok := ask(key, fn)
	return got, ok
}

// ridersRank is what a set of spend riders would do for this context's
// spell: -1 when any would fire, +1 when it carries riders the planner
// can judge and none would (the mana is HELD BACK), 0 otherwise — no
// riders, or none it can judge.
func (ctx ManaSpendContext) ridersRank(rs []ManaSpendRider, ask conditionAsker) int8 {
	if ctx.wish == nil || len(rs) == 0 {
		return 0
	}
	held := false
	for _, r := range rs {
		fires, known := ctx.riderFires(r, ask)
		if !known {
			continue
		}
		if fires {
			return -1
		}
		held = true
	}
	if held {
		return 1
	}
	return 0
}

// tokenRank ranks one floating mana for the pool solver, on the scale
// ridersRank uses: -1 spend it here first, +1 keep it for something it
// would do more for. A token made by a source of a kind the spell reads
// back (#1212: Hired Hexblade's Treasure) is ranked first even when it
// also carries a rider that would not fire.
func (ctx ManaSpendContext) tokenRank(tok ManaToken) int8 {
	if ctx.wish == nil {
		return 0
	}
	if ctx.wish.kinds != 0 && tok.SourceKinds.HasAny(ctx.wish.kinds) {
		return -1
	}
	return ctx.ridersRank(tok.Riders, ctx.verdictAsker())
}

// spendGenericWished is spendGeneric with #3853's preferences: mana that
// would do something for the spell first, then mana of a colour the spell
// reads, then everything else in the usual colourless-first order, and
// mana whose rider would be wasted last. With no wish, or a pool the wish
// does not distinguish, it is spendGeneric.
//
// Only the ORDER of runs the solver was already free to take changes: the
// same set is walked, so the cost is payable exactly when it was.
func spendGenericWished(pool ManaPool, used []int, order []int, need int, ctx ManaSpendContext) int {
	w := ctx.wish
	if w == nil || need <= 0 {
		return spendGeneric(pool, used, order, need)
	}
	var first, mid, last []int
	for _, i := range order {
		switch ctx.tokenRank(pool[i]) {
		case -1:
			first = append(first, i)
		case 1:
			last = append(last, i)
		default:
			mid = append(mid, i)
		}
	}
	if len(first) == 0 && len(last) == 0 && len(w.colors) == 0 {
		return spendGeneric(pool, used, order, need)
	}
	need = spendGeneric(pool, used, first, need)
	if len(w.colors) > 0 && need > 0 {
		var wanted, rest []int
		for _, i := range mid {
			if containsFold(w.colors, pool[i].Color) {
				wanted = append(wanted, i)
			} else {
				rest = append(rest, i)
			}
		}
		need = spendGeneric(pool, used, wanted, need)
		mid = rest
	}
	need = spendGeneric(pool, used, mid, need)
	return spendGeneric(pool, used, last, need)
}

// rankPoolOrder re-sorts a spendOrder result so that, among runs with the
// same number of restrictions, the ones that would do something for the
// spell come first and the ones whose rider would be wasted last. Stable,
// so a pool the wish does not distinguish keeps its insertion order.
func rankPoolOrder(pool ManaPool, order []int, ctx ManaSpendContext) {
	if ctx.wish == nil || len(order) < 2 {
		return
	}
	rank := make(map[int]int8, len(order))
	distinguishes := false
	for _, i := range order {
		r := ctx.tokenRank(pool[i])
		rank[i] = r
		if r != 0 {
			distinguishes = true
		}
	}
	if !distinguishes {
		return
	}
	sort.SliceStable(order, func(a, b int) bool {
		ra, rb := len(pool[order[a]].Restrictions), len(pool[order[b]].Restrictions)
		if ra != rb {
			return ra > rb
		}
		return rank[order[a]] < rank[order[b]]
	})
}

// autoTapWish is everything the planner is asked to prefer for one cast:
// #1212's source kinds, and the spend context whose wish (riders, colours)
// #3853 reads. The zero value prefers nothing and compares equal for every
// candidate, so a planner called with it is byte-identical to the
// planner before either feature existed.
//
// Local to one planning call, so it may hold the Game and a memo: it is
// not reachable from Game state.
type autoTapWish struct {
	kinds ManaSourceKinds
	spend ManaSpendContext

	g          *Game
	controller uuid.UUID
	card       *Card
	cardLooked bool
	memo       map[string]bool
}

// autoTapWishFor assembles the wish for a cast paid under `ctx`, with the
// caller's source kinds (WantedManaSourcesFor, or 0 for the enumerator's
// wishless question).
//
// Caller must hold g.mu.
func (g *Game) autoTapWishFor(controller uuid.UUID, ctx ManaSpendContext, kinds ManaSourceKinds) *autoTapWish {
	return &autoTapWish{kinds: kinds, spend: ctx, g: g, controller: controller}
}

// kindSet is #1212's source wish, zero for no wish.
func (w *autoTapWish) kindSet() ManaSourceKinds {
	if w == nil {
		return 0
	}
	return w.kinds
}

// ask answers a rider's Condition against the spell being cast, once.
func (w *autoTapWish) ask(key string, fn func(g *Game, controller uuid.UUID, paidFor Card) bool) (bool, bool) {
	if w == nil || w.g == nil || w.spend.wish == nil {
		return false, false
	}
	if got, ok := w.memo[key]; ok {
		return got, true
	}
	if !w.cardLooked {
		w.cardLooked = true
		if c, ok := w.g.LookupCardForEffect(w.spend.wish.paidID); ok {
			w.card = &c
		}
	}
	if w.card == nil {
		return false, false
	}
	if w.memo == nil {
		w.memo = map[string]bool{}
	}
	got := fn(w.g, w.controller, *w.card)
	w.memo[key] = got
	return got, true
}

// riderRank is what an ability's spend riders would do for the spell.
func (w *autoTapWish) riderRank(rs []ManaSpendRider) int8 {
	if w == nil || len(rs) == 0 {
		return 0
	}
	return w.spend.ridersRank(rs, w.ask)
}

// colors is the colours a "{X} was spent" card reads, nil for none.
func (w *autoTapWish) colors() []string {
	if w == nil || w.spend.wish == nil {
		return nil
	}
	return w.spend.wish.colors
}

// distinct reports a converge / sunburst reader.
func (w *autoTapWish) distinct() bool {
	return w != nil && w.spend.wish != nil && w.spend.wish.distinct
}

// stampReads sets ReadsColor on each candidate: it makes a colour the
// spell names, or — for a converge / sunburst spell — any colour at all.
func (w *autoTapWish) stampReads(sources []tapSource) {
	colors, distinct := w.colors(), w.distinct()
	if len(colors) == 0 && !distinct {
		return
	}
	for i := range sources {
		for _, c := range offeredColors(sources[i]) {
			if distinct || containsFold(colors, c) {
				sources[i].ReadsColor = true
				break
			}
		}
	}
}

// offeredColors is the distinct WUBRG colours a candidate can make, in
// slot order.
func offeredColors(s tapSource) []string {
	var out []string
	for _, slot := range s.Slots {
		for _, opt := range slot.Options {
			if isColorSymbol(opt) && !containsFold(out, opt) {
				out = append(out, opt)
			}
		}
	}
	return out
}

// colorSpread is the generic recruit's view of a colour reader (#3853):
// which colours the plan already makes, and which the spell wants more of.
// Nil for every spell that reads no colours, and every method is nil-safe
// and a no-op there.
//
// It is greedy and consults no budget: it re-ranks among candidates the
// order has already declared equal (sameGenericGroup), so it can neither
// change the candidate set nor reach for a tier the order kept back.
type colorSpread struct {
	distinct bool
	wished   []string
	have     map[string]bool
}

// newColorSpread builds the spread for a plan whose coloured pass has
// finished: `have` is the colours it already makes for sure — the
// single-colour pips of the cost and the single-option slots of the
// sources it booked. A dual land booked for a pip may make either colour,
// so it is not counted.
func newColorSpread(w *autoTapWish, cost ParsedCost, sources []tapSource, used []bool) *colorSpread {
	colors, distinct := w.colors(), w.distinct()
	if len(colors) == 0 && !distinct {
		return nil
	}
	cs := &colorSpread{distinct: distinct, wished: colors, have: map[string]bool{}}
	for _, req := range cost.Required {
		if len(req.Options) == 1 && isColorSymbol(req.Options[0]) {
			cs.have[req.Options[0]] = true
		}
	}
	for i := range sources {
		if !used[i] {
			continue
		}
		for _, slot := range sources[i].Slots {
			if len(slot.Options) == 1 && isColorSymbol(slot.Options[0]) {
				cs.have[slot.Options[0]] = true
			}
		}
	}
	return cs
}

// novel reports whether colour c is one the spell would like the plan to
// add: any colour it does not make yet for a converge / sunburst spell, a
// named colour it does not make yet for a card that names several
// (Moonhold wants both {R} and {W}).
func (cs *colorSpread) novel(c string) bool {
	return !cs.have[c] && (cs.distinct || containsFold(cs.wished, c))
}

// spreads reports whether the spell counts HOW MANY of its colours were
// spent, which is when re-ranking within a group can buy anything. A card
// that names one colour is served by the static ReadsColor ordering alone.
func (cs *colorSpread) spreads() bool {
	return cs != nil && (cs.distinct || len(cs.wished) > 1)
}

// newColors counts the colours `s` could add that the spell would like
// the plan to add, capped at the number of mana it makes (one colour per
// slot).
func (cs *colorSpread) newColors(s tapSource) int {
	n := 0
	for _, c := range offeredColors(s) {
		if cs.novel(c) {
			n++
		}
	}
	return min(n, len(s.Slots))
}

// refine returns the candidate the recruit should take: `pick`, unless a
// converge / sunburst spell can have a source of the same group that adds
// more colours to the plan and still fits the deficit the way `pick`
// does. Ties keep the order's choice.
func (cs *colorSpread) refine(sources []tapSource, order []int, eligible func(int) bool, pick, deficit int) int {
	if !cs.spreads() {
		return pick
	}
	best, bestNew := pick, cs.newColors(sources[pick])
	overshoot := len(sources[pick].Slots) > deficit
	for _, i := range order {
		if i == pick || !eligible(i) || !sameGenericGroup(sources[pick], sources[i]) {
			continue
		}
		if !overshoot && len(sources[i].Slots) > deficit {
			continue
		}
		if n := cs.newColors(sources[i]); n > bestNew {
			best, bestNew = i, n
		}
	}
	return best
}

// colorFor is the colour to make with `s` and books it as made: the first
// colour a converge spell has not got yet, or the first colour a "{X} was
// spent" card named that `s` offers. Empty when the spell reads nothing
// of this source, which leaves the executor's own pick.
func (cs *colorSpread) colorFor(s tapSource) string {
	if cs == nil {
		return ""
	}
	offered := offeredColors(s)
	for _, c := range offered {
		if cs.novel(c) {
			cs.have[c] = true
			return c
		}
	}
	for _, c := range cs.wished {
		if containsFold(offered, c) {
			cs.have[c] = true
			return c
		}
	}
	return ""
}

// pickColorForPlanned is pickColorForSlot for a planned entry: a pending
// coloured requirement the slot can pay is paid first, exactly as before,
// and only a slot that would fall through to its first option makes the
// colour the planner recruited it for (plannedTap.Prefer) when it offers
// it. A preference that could leave a pip unpaid would be a change to
// payability; this one cannot, because it runs only when no pip matched.
func pickColorForPlanned(planned plannedTap, options []string, pending *[]ColorRequirement) string {
	if len(options) == 0 {
		return ""
	}
	if best, bestOpt := mostRestrictiveRequirement(options, *pending); best >= 0 {
		*pending = append((*pending)[:best], (*pending)[best+1:]...)
		return bestOpt
	}
	if planned.Prefer != "" && colorOffered(options, planned.Prefer) {
		return planned.Prefer
	}
	return options[0]
}

// wantsColorList normalises a declaration's colours to the uppercase
// WUBRG alphabet, dropping anything else.
func wantsColorList(in []string) []string {
	var out []string
	for _, c := range in {
		c = strings.ToUpper(strings.TrimSpace(c))
		if isColorSymbol(c) && !containsFold(out, c) {
			out = append(out, c)
		}
	}
	return out
}
