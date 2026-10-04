package legal

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cuts.go — ADR 0122 §6.2: the enumerator reports what it leaves out.
//
// Every count cap in this package is a cost bound or a policy
// judgement ("what is worth naming"), never a rule of the game, so a
// list cut by one is a list of SOME of a seat's legal moves. Until
// ADR 0122 nothing said so. Each cap site now files a Cut against the
// card or prompt it was expanding, and EnumerateReport hands the cuts
// back beside the moves. EnumerateFor and EnumerateLocked still return
// the moves alone: the in-process bot keeps its caps and its list
// exactly as before (ADR 0122 §6.2, "the runner does not use this").
//
// A caller that wants the rest asks again for that one card or prompt
// with Options.Source or Options.Choice, which lifts every count cap to
// ExpandCeiling. What is still cut then is reported again, under
// CapCeiling or the cap that bound it, and never hidden.

// Cap names one of the enumerator's count caps. The values are wire
// tokens (docs/protocol.md, "The cut report"): stable once shipped.
type Cap string

const (
	// CapPerSource is Options.MaxExpansionPerSource: moves per card
	// (or per prompt) across its targets, modes and cost choices.
	CapPerSource Cap = "per_source"
	// CapMaxX is Options.MaxX: the largest X the enumerator will try.
	CapMaxX Cap = "max_x"
	// CapVariableCounts is maxEnumeratedVariableCounts: how many
	// different counts a "sacrifice any number" style cost is offered.
	CapVariableCounts Cap = "variable_counts"
	// CapSubsetScan is subsetScanBudget: how many candidate subsets a
	// filtered card-set prompt tests before it gives up.
	CapSubsetScan Cap = "subset_scan"
	// CapCreatureTypes is creatureTypeAnswersCap: the creature types
	// offered for a choose-a-creature-type prompt.
	CapCreatureTypes Cap = "creature_types"
	// CapCardNames is cardNameAnswersCap: the card names offered for a
	// choose-a-card-name prompt.
	CapCardNames Cap = "card_names"
	// CapCostPayments is maxEnumeratedCostPayments: the ways one
	// alternative cost's card component is offered to be paid.
	CapCostPayments Cap = "cost_payments"
	// CapRepeats is maxEnumeratedRepeats: how many times a repeatable
	// optional cost (multikicker) is offered.
	CapRepeats Cap = "repeats"
	// CapCeiling is ExpandCeiling, the hard ceiling on what one
	// expanded request may return.
	CapCeiling Cap = "ceiling"
)

// ExpandCeiling is what every count cap is raised to when Options names
// one card or one prompt (ADR 0122 §6.2), and the most moves such a
// request returns. Creature types are a vocabulary of about 300, so
// under it they are listed in full.
const ExpandCeiling = 512

// CleanupDiscardChoice is the Options.Choice — and the Cut.Choice —
// that names the cleanup-step discard (CR 514.1), which is a decision
// the seat owes without a pending-choice ID of its own.
const CleanupDiscardChoice = "cleanup_discard"

// Cut is one place a cap stopped the enumeration with work left.
//
// Source is the card being expanded, uuid.Nil for a prompt with none.
// Choice is the pending choice's ID (or CleanupDiscardChoice) when the
// cut was in answering one, "" otherwise. A cut can carry both: a
// prompt raised by a card.
//
// Omitted is how many candidate moves the cap kept the enumerator from
// building: each one legal as far as the cap site could tell, though a
// later check (an affordability probe the site never reached) might
// still have refused some. It is exact when AtLeast is false. When
// AtLeast is true it is a lower bound, because the site stopped before
// it could count the rest without doing the work the cap exists to
// save. A lower bound of 0 means the walk stopped with candidates left
// untested and cannot say whether any of them would have been legal.
type Cut struct {
	Source  uuid.UUID
	Choice  string
	Cap     Cap
	Omitted int
	AtLeast bool
}

// Report is one enumeration and what it left out.
type Report struct {
	Moves []Move
	Cuts  []Cut
}

// EnumerateReport is EnumerateForWithOptions with the cut report.
// Callers must not hold g.mu.
func EnumerateReport(g *game.Game, seat uuid.UUID, opts Options) Report {
	if g == nil || seat == uuid.Nil {
		return Report{}
	}
	var out Report
	g.ReadSnapshot(func() {
		out = enumerateReportLocked(g, seat, opts.withDefaults())
	})
	return out
}

// EnumerateReportLocked is EnumerateReport for a caller that already
// holds g's read lock — see EnumerateLocked.
func EnumerateReportLocked(g *game.Game, seat uuid.UUID, opts Options) Report {
	if g == nil || seat == uuid.Nil {
		return Report{}
	}
	return enumerateReportLocked(g, seat, opts.withDefaults())
}

// scope is what the enumerator is expanding right now: the card a cap
// site files its cut against, and what the Source / Choice filter is
// asked about.
type scope struct {
	source uuid.UUID
	choice string
}

// expanding reports whether the caller asked for one card or one
// prompt, which is what lifts the caps (Options doc).
func (o Options) expanding() bool {
	return o.Source != uuid.Nil || o.Choice != ""
}

// enter makes s the current scope and reports whether the filter admits
// it. With no filter every scope is admitted, so the walk is exactly the
// walk it always was. A caller that gets false skips the work: nothing
// it would add is wanted.
func (e *enumerator) enter(s scope) bool {
	e.scope = s
	e.admitted = e.admits(s)
	return e.admitted
}

// leave clears the scope, for a move that belongs to no card (the pass,
// finish_blocks). Under a filter such a move is never wanted.
func (e *enumerator) leave() {
	e.scope = scope{}
	e.admitted = !e.opts.expanding()
}

func (e *enumerator) admits(s scope) bool {
	if !e.opts.expanding() {
		return true
	}
	if e.opts.Choice != "" {
		return s.choice == e.opts.Choice
	}
	return s.source != uuid.Nil && s.source == e.opts.Source
}

// noteCut files a cut against the current scope. Two cuts of one cap in
// one scope are one entry, their counts summed: a card whose budget ran
// out in three places left out at least the three candidates it was
// holding. A cut outside the filter is not the caller's business.
func (e *enumerator) noteCut(c Cap, omitted int, atLeast bool) {
	if !e.report || !e.admitted {
		return
	}
	if omitted < 0 {
		omitted = 0
	}
	for i := range e.cuts {
		k := &e.cuts[i]
		if k.Source == e.scope.source && k.Choice == e.scope.choice && k.Cap == c {
			if c == CapMaxX {
				// One card's X is searched once per announcement it
				// prices, and every search past the cap is the SAME
				// larger X left out. Never summed.
				k.Omitted = max(k.Omitted, omitted)
				k.AtLeast = k.AtLeast || atLeast
				return
			}
			k.Omitted = satAdd(k.Omitted, omitted)
			k.AtLeast = k.AtLeast || atLeast
			return
		}
	}
	e.cuts = append(e.cuts, Cut{
		Source:  e.scope.source,
		Choice:  e.scope.choice,
		Cap:     c,
		Omitted: omitted,
		AtLeast: atLeast,
	})
}

// countCap bounds an exact count, so a binomial over a large pool never
// overflows. A count at the bound is reported AtLeast.
const countCap = 1 << 30

func satAdd(a, b int) int {
	if a >= countCap-b {
		return countCap
	}
	return a + b
}

// binomial is C(n, k), saturating at countCap.
func binomial(n, k int) int {
	if k < 0 || k > n {
		return 0
	}
	if k > n-k {
		k = n - k
	}
	r := 1
	for i := 1; i <= k; i++ {
		// r * (n-k+i) / i is exact at every step.
		if r > countCap/(n-k+i) {
			return countCap
		}
		r = r * (n - k + i) / i
	}
	return r
}

// subsetCount is how many subsets of a pool of n have a size in lo..hi.
func subsetCount(n, lo, hi int) int {
	if hi > n {
		hi = n
	}
	if lo < 0 {
		lo = 0
	}
	total := 0
	for k := lo; k <= hi; k++ {
		total = satAdd(total, binomial(n, k))
	}
	return total
}

// combos is combinations() with its cut filed: the pool's subsets are
// counted, not walked, so the count is exact and costs nothing the cap
// was saving.
func (e *enumerator) combos(pool []uuid.UUID, lo, hi, limit int, c Cap) [][]uuid.UUID {
	out := combinations(pool, lo, hi, limit)
	if e.report && len(out) >= limit && limit > 0 {
		if total := subsetCount(len(pool), lo, hi); total > len(out) {
			e.noteCut(c, total-len(out), total >= countCap)
		}
	}
	return out
}

// budgetSpent files the cut for an expansion budget that ran out while
// a candidate was in hand: the candidate is one move left out, and the
// walk does not go on to count the others.
func (e *enumerator) budgetSpent() {
	e.noteCut(CapPerSource, 1, true)
}

// noteMaxX files the cut for an X search the MaxX cap stopped: payable
// reports whether a given X could be paid, and is asked once, for the
// first X past the cap, only when the search reached it.
func (e *enumerator) noteMaxX(best int, payable func(x int) bool) {
	if e.report && best == e.opts.MaxX && payable(e.opts.MaxX+1) {
		e.noteCut(CapMaxX, 1, true)
	}
}

// capOr is a count cap's value: its default, or ExpandCeiling when the
// caller asked for one card or prompt.
func (e *enumerator) capOr(def int) int {
	if e.opts.expanding() {
		return ExpandCeiling
	}
	return def
}

// ceilingFull reports whether an expanded request has already returned
// ExpandCeiling moves, filing the cut for the move it refuses.
func (e *enumerator) ceilingFull() bool {
	if !e.opts.expanding() || e.added < ExpandCeiling {
		return false
	}
	e.noteCut(CapCeiling, 1, false)
	return true
}

func enumerateReportLocked(g *game.Game, seat uuid.UUID, opts Options) Report {
	e := newEnumerator(g, seat, opts)
	if e == nil {
		return Report{}
	}
	e.report = true
	e.run()
	return Report{Moves: e.out, Cuts: e.cuts}
}

// refCombos is combinationsRefs with its cut filed. The walk is asked
// for one set past the limit, which leaves the first `limit` exactly as
// they were: a set beyond them means the cap cut. Without a set rule
// the sets are counted, not walked, and the count is exact; with one,
// the walk stops at the first spare set and the count is a lower bound.
func (e *enumerator) refCombos(cands []game.TargetRef, lo, hi, limit int, keys setRuleKeys) [][]game.TargetRef {
	if !e.report {
		return combinationsRefs(cands, lo, hi, limit, keys)
	}
	out := combinationsRefs(cands, lo, hi, limit+1, keys)
	if len(out) <= limit {
		return out
	}
	out = out[:limit]
	e.noteSetCut(len(cands), lo, hi, limit, keys)
	return out
}

// noteSetCut files the per-source cut for a target-set walk that had a
// set past its limit.
func (e *enumerator) noteSetCut(n, lo, hi, limit int, keys setRuleKeys) {
	if keys.different == nil && keys.same == nil {
		if total := subsetCount(n, lo, hi); total > limit {
			e.noteCut(CapPerSource, total-limit, total >= countCap)
			return
		}
	}
	e.noteCut(CapPerSource, 1, true)
}

// modePicks is game.ModePickSelections at this enumeration's cap, with
// its cut filed. The selections come out in a fixed order and stop at
// the budget, so asking for one more leaves the first ones as they were.
func (e *enumerator) modePicks(c *game.PendingChoice) [][]int {
	limit := e.opts.MaxExpansionPerSource
	if !e.report {
		return game.ModePickSelections(c, limit)
	}
	out := game.ModePickSelections(c, limit+1)
	if len(out) > limit {
		out = out[:limit]
		e.noteCut(CapPerSource, 1, true)
	}
	return out
}

// noteXCountedTargets files the cut for an X-counted target step opened
// to MaxX targets (a per-target price, cast.go): a step with more legal
// targets than that had larger announcements the cap closed. Asked of
// the engine only for a caller that wants the report.
func (e *enumerator) noteXCountedTargets(src game.TargetSource, steps []game.AnnouncedClause, xSteps []int) {
	if !e.report {
		return
	}
	for _, i := range xSteps {
		clause := steps[i].Clause
		lt := e.g.LegalTargetsForEffect(src, &clause)
		if n := len(lt.Players) + len(lt.Cards); n > e.opts.MaxX {
			e.noteCut(CapMaxX, n-e.opts.MaxX, true)
		}
	}
}
