package game

import (
	"slices"
	"strconv"
	"strings"
)

// coloured_pass.go — the pool solver's coloured pass, worked run by run
// (#3904).
//
// The pool has been run-length since #3041: a million {B} is one
// ManaToken with a count. The cost side has one long run too: a folded
// "spend only black mana on X" (spend_only.go) is X identical
// requirements in a row. The pass used to expand both into single mana
// ("units") and match them one at a time, which at the X bound
// (MaxAnnouncedX) was several hundred megabytes per payment. It now
// works on the runs themselves:
//
//   - the requirements are grouped into reqRuns, n identical
//     requirements each, so the long X run is one entry (since #3909
//     the fold writes it as one run to begin with);
//   - each pool run keeps how many of its mana are booked and, run-length
//     too, which requirement run each booked mana belongs to (ownerSeg);
//   - what the pass leaves unpaid, or defers to the generic pass, comes
//     back as reqRuns, so MissingFor and the paper-payment picker can
//     name a long shortfall once.
//
// It is EXACT, not an approximation: the bookings, and so the mana every
// payment spends, are the ones the unit-by-unit pass made. The unit pass
// had three properties this one keeps by construction:
//
//   - the booked mana of a run are always its first ones (the walk takes
//     the first unbooked mana in order, and a booking is moved, never
//     dropped), so "booked" is a count;
//   - the augmenting path visits a run's mana in order, so what it has
//     visited is a count too;
//   - identical requirements behave identically: the walk books them in
//     order into the first free mana, and an augmenting search that
//     fails for one fails for all of them, booking nothing.
//
// The third is what lets a run of requirements be booked in one step,
// and lets a run of booked mana whose owner cannot move be skipped in
// one step rather than visited mana by mana.

// reqRun is n identical colour requirements in a row: the counted form
// the coloured pass reads and returns.
type reqRun struct {
	req ColorRequirement
	n   int
	// x marks the run that is the mana paid for {X}, the part a "spend
	// only <colours> mana on X" clause folded into coloured
	// requirements (#3941, spend_only.go). It keeps that run apart from
	// a printed requirement with the same options, so the pass can say
	// which mana paid X; it changes nothing about how the run is paid.
	x bool
}

// sameRequirement reports whether a and b are the same requirement in
// every field. TestSameRequirementReadsEveryField keeps it in step with
// ColorRequirement.
func sameRequirement(a, b ColorRequirement) bool {
	return a.Phyrexian == b.Phyrexian && a.NumericAlt == b.NumericAlt &&
		a.HasNumericAlt == b.HasNumericAlt && a.Snow == b.Snow &&
		a.AnyMana == b.AnyMana && a.LifeGranted == b.LifeGranted &&
		slices.Equal(a.Options, b.Options)
}

// appendReqRun appends n of r, merging them into the last run when it
// is the same requirement.
func appendReqRun(runs []reqRun, r ColorRequirement, n int) []reqRun {
	return appendTaggedReqRun(runs, reqRun{req: r, n: n})
}

// appendTaggedReqRun is appendReqRun for a run that may be the X part.
// It merges only into a run with the same tag, so the X part stays its
// own run even beside a printed requirement of the same colours.
func appendTaggedReqRun(runs []reqRun, r reqRun) []reqRun {
	if r.n <= 0 {
		return runs
	}
	if k := len(runs); k > 0 && runs[k-1].x == r.x && sameRequirement(runs[k-1].req, r.req) {
		runs[k-1].n = satAddMana(runs[k-1].n, r.n)
		return runs
	}
	return append(runs, r)
}

// requirementRuns is reqs followed by the counted run `tail` (a
// cost's folded run, #3909), as runs, in widenedLast's order (every
// slot a spend grant widened to any mana behind the ones that still
// name their colours), without copying reqs or writing the tail out.
func requirementRuns(reqs []ColorRequirement, tail reqRun) []reqRun {
	var out []reqRun
	widened, twobrid := false, false
	for _, r := range reqs {
		if r.AnyMana {
			widened = true
			continue
		}
		if r.isTwobrid() {
			twobrid = true
			continue
		}
		out = appendReqRun(out, r, 1)
	}
	if !tail.req.AnyMana {
		out = appendTaggedReqRun(out, tail)
	}
	// #4177: a two-mana hybrid ({2/B}) takes its colour only after every
	// plain symbol has had its pick, so it never holds the one mana a
	// plain {B} needed when it could have been paid with two others.
	if twobrid {
		for _, r := range reqs {
			if r.isTwobrid() && !r.AnyMana {
				out = appendReqRun(out, r, 1)
			}
		}
	}
	if widened {
		for _, r := range reqs {
			if r.AnyMana {
				out = appendReqRun(out, r, 1)
			}
		}
	}
	if tail.req.AnyMana {
		out = appendTaggedReqRun(out, tail)
	}
	return out
}

// isTwobrid reports whether the slot is a two-mana hybrid ({2/B}): one
// mana of its colour or N mana of any kind (CR 107.4e). The solvers pay
// the colour when they can and the N generic when they cannot
// (foldUnpaidTwobrid, solveColored).
func (r ColorRequirement) isTwobrid() bool {
	return r.HasNumericAlt && r.NumericAlt > 1 && !r.Phyrexian
}

// foldUnpaidTwobrid takes the two-mana hybrids out of a pass's unpaid
// requirements and returns how much generic mana they owe instead (#4177,
// CR 107.4e, 601.2b): a {2/B} no black mana paid is two of anything. The
// rest come back unchanged, still unpaid.
func foldUnpaidTwobrid(unpaid []reqRun) (rest []reqRun, generic int) {
	for _, r := range unpaid {
		if r.req.isTwobrid() {
			generic = satAddMana(generic, r.n*r.req.NumericAlt)
			continue
		}
		rest = append(rest, r)
	}
	return rest, generic
}

// requirementRuns is the cost's coloured requirements as runs: the
// listed ones and the folded run, in widenedLast's order.
func (c ParsedCost) requirementRuns() []reqRun {
	return requirementRuns(c.Required, c.folded)
}

// coalesced is runs with adjacent runs of the same requirement joined,
// whether or not one is the X part (#3941). For a breakdown that names
// symbols to a player: "{B}×1000001", not "{B}" and "{B}×1000000".
func coalesced(runs []reqRun) []reqRun {
	var out []reqRun
	for _, r := range runs {
		r.x = false
		out = appendTaggedReqRun(out, r)
	}
	return out
}

// reqRunsTotal is how many requirements the runs stand for.
func reqRunsTotal(runs []reqRun) int {
	n := 0
	for _, r := range runs {
		n = satAddMana(n, r.n)
	}
	return n
}

// lastOfReqRuns is the last k requirements of runs, in order.
func lastOfReqRuns(runs []reqRun, k int) []reqRun {
	var out []reqRun
	for i := len(runs) - 1; i >= 0 && k > 0; i-- {
		take := min(k, runs[i].n)
		out = append(out, reqRun{req: runs[i].req, n: take, x: runs[i].x})
		k -= take
	}
	slices.Reverse(out)
	return out
}

// maxListedRun is the longest run of one symbol a breakdown names one
// symbol at a time: MissingFor's missing list, the paper-payment
// picker's slots and the log's paper-payment symbols. A longer run is
// named once, with its count (symbolRun), so a shortfall the size of an
// announced X is one entry rather than X of them.
const maxListedRun = maxListedGeneric

// appendSymbolRun appends n of sym to out: n copies up to maxListedRun,
// and past it one "sym×n".
func appendSymbolRun(out []string, sym string, n int) []string {
	if n > maxListedRun {
		return append(out, symbolRun(sym, n))
	}
	for range n {
		out = append(out, sym)
	}
	return out
}

// writeSymbolRun is appendSymbolRun into a string.
func writeSymbolRun(b *strings.Builder, sym string, n int) {
	if n > maxListedRun {
		b.WriteString(symbolRun(sym, n))
		return
	}
	for range n {
		b.WriteString(sym)
	}
}

// symbolRun names a long run of one symbol: "{B}×500000".
func symbolRun(sym string, n int) string {
	return sym + "×" + strconv.Itoa(n)
}

// ownerSeg is n booked mana of a pool run, all booked for the
// requirement run ri.
type ownerSeg struct {
	ri int
	n  int
}

// colourRun is one pool run as the coloured pass sees it.
type colourRun struct {
	pool  int    // index into the pool
	color string // the run's mana
	free  int    // mana the pass may book: N less what was already used
	// booked is how many of the run's first mana are booked, and
	// owners which requirement run each belongs to, in order.
	booked int
	owners []ownerSeg
	// visited is how many of the run's first mana the current
	// augmenting search has visited.
	visited int
}

// bookN books the run's next n free mana for ri.
func (c *colourRun) bookN(ri, n int) {
	if k := len(c.owners); k > 0 && c.owners[k-1].ri == ri {
		c.owners[k-1].n += n
	} else {
		c.owners = append(c.owners, ownerSeg{ri: ri, n: n})
	}
	c.booked += n
}

// segAt is the owner segment holding booked mana u, and the offset of
// its first mana.
func (c *colourRun) segAt(u int) (si, start int) {
	for i, s := range c.owners {
		if u < start+s.n {
			return i, start
		}
		start += s.n
	}
	panic("game: colourRun.segAt past the booked mana")
}

// take books mana u for ri: the next free mana when u is the first
// unbooked one, and otherwise a booked mana whose owner has moved.
func (c *colourRun) take(u, ri int) {
	if u >= c.booked {
		c.bookN(ri, 1)
		return
	}
	si, start := c.segAt(u)
	s := c.owners[si]
	if s.ri == ri {
		return
	}
	off := u - start
	repl := make([]ownerSeg, 0, 3)
	if off > 0 {
		repl = append(repl, ownerSeg{ri: s.ri, n: off})
	}
	repl = append(repl, ownerSeg{ri: ri, n: 1})
	if rest := s.n - off - 1; rest > 0 {
		repl = append(repl, ownerSeg{ri: s.ri, n: rest})
	}
	c.owners = slices.Replace(c.owners, si, si+1, repl...)
	// Keep the segments maximal, so a run's owners stay a handful.
	out := c.owners[:0]
	for _, s := range c.owners {
		if k := len(out); k > 0 && out[k-1].ri == s.ri {
			out[k-1].n += s.n
			continue
		}
		out = append(out, s)
	}
	c.owners = out
}

// colourSolver is one coloured pass's state.
type colourSolver struct {
	reqs []reqRun
	runs []colourRun
	// failed marks, within one augmenting search, the requirement runs
	// whose search has already failed (see augment).
	failed []bool
	// favor is the colours the X part takes first (see colouredPassX).
	favor []string
}

// colouredPass pays each requirement in `reqs` with one mana, booking
// it in `used`, and returns the widened slots it left to the generic
// pass and the requirements it could not pay, both as runs in `reqs`
// order.
//
// Each requirement names a set of allowed colors (monocolored = one
// entry, hybrid = two). It takes the first available mana whose color
// matches. Phyrexian / numeric-alt requirements pay the color half
// here (the life half is strikePhyrexianLifeLocked's).
//
// A WIDENED slot (AnyMana — a spend grant's, #1589 / #1600) takes a
// mana of a colour it PRINTS here, like any other slot, and otherwise
// joins the generic demand, which any mana pays. That is the same
// answer to "payable?" as admitting any mana here, and a better answer
// to "with what?": the colour the card asked for is spent when the
// pool has it, so a grant the player "may" use never costs a Firespout
// its red mode.
//
// ADR 0118 §1: a requirement the first-match walk leaves unpaid is
// given one more chance by an augmenting path (reassign) before it is
// called unpaid. A payment the walk already makes is made with exactly
// the same mana as before.
//
// It costs O(pool runs × requirement runs) plus the augmenting paths,
// however much mana the pool holds and however long the X (#3041,
// #3904); see the file comment.
func colouredPass(pool ManaPool, used []int, order []int, reqs []reqRun) (deferred, unpaid []reqRun) {
	deferred, unpaid, _ = colouredPassX(pool, used, order, reqs, nil)
	return deferred, unpaid
}

// colouredPassX is colouredPass that also reports which mana paid the
// X part (#3941): xUsed[i] is how many of pool run i's mana went to the
// requirement runs marked x, or nil when no run is marked. The bookings
// are the same ones colouredPass makes.
//
// `favor` is the colours of mana the X part takes FIRST when it has a
// choice (a spell that reads a colour spent on X: Soul Burn's {B}). It
// reorders which of the runs the X part's walk takes from and nothing
// else: every other requirement is walked as before and the augmenting
// path still repairs any refusal, so whether the cost is payable never
// depends on it. The X part is the last requirement walked, so the
// printed coloured symbols have already taken what they need.
func colouredPassX(pool ManaPool, used []int, order []int, reqs []reqRun, favor []string) (deferred, unpaid []reqRun, xUsed []int) {
	if len(reqs) == 0 {
		return nil, nil, nil
	}
	s := colourSolver{reqs: reqs, runs: make([]colourRun, 0, len(order)), favor: favor}
	for _, i := range order {
		if free := pool[i].N() - used[i]; free > 0 {
			s.runs = append(s.runs, colourRun{pool: i, color: pool[i].Color, free: free})
		}
	}
	for ri, rr := range reqs {
		m := s.walk(ri, rr.n)
		if m > 0 && !rr.req.AnyMana {
			m -= s.reassign(ri, m)
		}
		if m == 0 {
			continue
		}
		if rr.req.AnyMana {
			deferred = append(deferred, reqRun{req: rr.req, n: m, x: rr.x})
		} else {
			unpaid = append(unpaid, reqRun{req: rr.req, n: m, x: rr.x})
		}
	}
	for _, c := range s.runs {
		used[c.pool] += c.booked
		for _, o := range c.owners {
			if reqs[o.ri].x {
				if xUsed == nil {
					xUsed = make([]int, len(pool))
				}
				xUsed[c.pool] += o.n
			}
		}
	}
	return deferred, unpaid, xUsed
}

// walk is the first-match walk for m requirements of run ri: each takes
// the first free mana, in order, of a colour the requirement PRINTS
// (matchColor over Options; a widened slot is asked the same question,
// and what it admits beyond its printed colours is paid in the generic
// pass instead, #1600). The m requirements are identical, so they fill
// the matching runs in order. Returns how many it could not book.
func (s *colourSolver) walk(ri, m int) int {
	opts := s.reqs[ri].req.Options
	if s.reqs[ri].x && len(s.favor) > 0 {
		// The X part takes the favoured colours' mana before any other:
		// the same runs, in another order (colouredPassX).
		for _, favoured := range [2]bool{true, false} {
			for k := range s.runs {
				if m == 0 {
					break
				}
				c := &s.runs[k]
				if !matchColor(c.color, opts) || containsFold(s.favor, c.color) != favoured {
					continue
				}
				if t := min(m, c.free-c.booked); t > 0 {
					c.bookN(ri, t)
					m -= t
				}
			}
		}
		return m
	}
	for k := range s.runs {
		if m == 0 {
			break
		}
		c := &s.runs[k]
		if !matchColor(c.color, opts) {
			continue
		}
		if t := min(m, c.free-c.booked); t > 0 {
			c.bookN(ri, t)
			m -= t
		}
	}
	return m
}

// reassign finds mana for up to m requirements of run ri that the walk
// left unpaid, one augmenting path each, and returns how many it found.
//
// An augmenting path moves mana already booked for an earlier
// requirement onto other mana that requirement also prints, so the pass
// pays every requirement whenever some assignment of mana to
// requirements could (CR 601.2h: the player chooses which mana pays
// which symbol, and would choose an assignment that works). It runs
// only after the walk has failed, so it repairs a refusal and never
// changes a success. ADR 0118 §1's top-up is the case it exists for: a
// {G} that floated before the auto-tapper ran sits in front of the
// plan's own mana, so the walk alone pays {G/W}{G} out of [{G}, {W}] by
// spending the {G} on the hybrid and then finds no {G} for the second
// symbol.
//
// Widened slots (AnyMana) never come here: they are walked last and
// defer to the generic pass rather than fail, so no booking this moves
// can belong to one.
//
// The first search that fails ends it: a failed search books nothing,
// so the next requirement of the run would search the same bookings and
// fail too.
func (s *colourSolver) reassign(ri, m int) int {
	if s.failed == nil {
		s.failed = make([]bool, len(s.reqs))
	}
	got := 0
	for got < m {
		for k := range s.runs {
			s.runs[k].visited = 0
		}
		clear(s.failed)
		k, u, ok := s.augment(ri)
		if !ok {
			break
		}
		s.runs[k].take(u, ri)
		got++
	}
	return got
}

// augment is one augmenting search for a requirement of run r: the
// mana now free for it, as (pool-run position, mana offset), with every
// booking along the path already moved. The caller books it.
//
// It visits the runs in order and each run's mana in order. A free mana
// ends the search. A booked one is freed by moving its owner, unless
// the owner is a widened slot, or prints the same colours as r: any
// mana such an owner could move to, this search reaches next with the
// same mana visited, and taking it here books the same mana, while
// following it would rescan the pool once per level (#3895).
//
// Two steps take a whole owner segment at once, both exact:
//
//   - a segment whose owner is not moved is skipped, each of its mana
//     being visited and passed over in turn;
//   - once moving an owner has failed, the rest of its segment is
//     skipped too. A failed search for a requirement run visits every
//     mana of a colour it prints, so asking it again finds nothing and
//     moves nothing. `failed` records that, so a requirement run whose
//     search has failed is not searched again within this augment.
func (s *colourSolver) augment(r int) (int, int, bool) {
	if s.failed[r] {
		return 0, 0, false
	}
	opts := s.reqs[r].req.Options
	for k := range s.runs {
		c := &s.runs[k]
		if !matchColor(c.color, opts) {
			continue
		}
		for c.visited < c.free {
			u := c.visited
			if u >= c.booked {
				c.visited++
				return k, u, true
			}
			si, start := c.segAt(u)
			seg := c.owners[si]
			end := start + seg.n
			prev := s.reqs[seg.ri].req
			if prev.AnyMana || slices.Equal(prev.Options, opts) {
				c.visited = end
				continue
			}
			c.visited++
			if k2, u2, ok := s.augment(seg.ri); ok {
				s.runs[k2].take(u2, seg.ri)
				return k, u, true
			}
			c.visited = max(c.visited, end)
		}
	}
	s.failed[r] = true
	return 0, 0, false
}
