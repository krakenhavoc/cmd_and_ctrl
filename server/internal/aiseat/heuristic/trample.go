package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// trample.go is the optimistic half of trample (#1504): how much of a
// blocked trampler's damage still reaches the player, as the attacker
// counts it when it asks "does this swing kill them?".
//
// The pessimistic half — "can their tramplers kill me?" — is
// crackBackPower in race.go, and it stays as it is: there a trampler
// is held only by a blocker that absorbs all of it. Here the question
// runs the other way, so every simplification has to err toward LESS
// damage: an estimate that says lethal when the defender can block its
// way out sends the bot into an alpha strike that loses the board.

// trampleAbsorb is how much of trampler a's damage one blocker b soaks
// up before the rest tramples over, as the engine assigns it.
//
// The engine's rule (CR 702.19b): each blocker is assigned lethal
// damage and the remainder may go to the player. Lethal is the
// blocker's toughness less the damage already marked on it, and 1
// from a deathtouch source (CR 702.19c, 702.2c). The single-blocker
// path in assignAndDealCombatDamageLocked floors lethal at 0; the
// canonical multi-blocker answer (legal.canonicalDamageAssignment)
// floors it at 1. Flooring at 1 here is the larger of the two, so the
// estimate never counts a point the engine would not send over.
//
// Two things the engine's assignment does not see, and this does:
//   - A first-strike blocker can kill the trampler before it deals any
//     damage at all, and several of them can do it together even when
//     none could alone. So a first striker (against an attacker
//     without first strike) is counted as holding the trampler in
//     full.
//   - Protection and indestructible do not change the assignment —
//     lethal is still toughness, and the overflow still goes over —
//     so they are deliberately not read here. `kills` reads them,
//     because they decide whether the blocker dies, which is a
//     different question.
func trampleAbsorb(a, b *protocol.CardView) int {
	if firstStrikes(b) && !firstStrikes(a) {
		return a.Power
	}
	if hasKeyword(a, "deathtouch") {
		return 1
	}
	if n := effectiveToughness(b); n > 1 {
		return n
	}
	return 1
}

// firstStrikes reports whether a creature deals damage in the
// first-strike combat damage step (CR 510.4).
func firstStrikes(c *protocol.CardView) bool {
	return hasKeyword(c, "first strike") || hasKeyword(c, "double strike")
}

// overflowPast is what attacker a still connects for after blocker b is
// added to its blocks, when `left` was getting through before: all of
// it stops for an attacker without trample, and a trampler loses only
// what b absorbs.
func overflowPast(a, b *protocol.CardView, left int) int {
	if !hasKeyword(a, "trample") {
		return 0
	}
	if over := left - trampleAbsorb(a, b); over > 0 {
		return over
	}
	return 0
}

// trampleBound is a lower bound on the damage a swing connects for
// when the defender blocks as well as it can and every trampler
// assigns what its blockers do not absorb to the player. Only a LOWER
// bound — the exact answer is a partition problem (which blockers go
// on which trampler) — and a lower bound is the direction an
// attacker's estimate may err.
//
// The defender's blocks save, in total, the power of every
// non-trampler it blocks (one blocker each, as in unblockedPower's
// matching) plus, for each trampler of power p, min(p, Σ credit) over
// that trampler's blockers, where a blocker's credit is
// min(p, trampleAbsorb). Sort one trampler's blockers by credit,
// biggest first: the j-th adds min(c_j, p − c_1 − … − c_{j−1}), and
// since every earlier credit is at least c_j, that is at most
//
//	slotValue(p, j, c_j) = min(c_j, p − (j−1)·c_j)
//
// — a number that depends on the blocker and the slot alone. So each
// trampler gets slots 1, 2, 3 … of one blocker each, the j-th worth
// slotValue to whoever fills it, and the defender fills them and the
// non-tramplers as it likes. Its real blocks are one way to fill them,
// so the best filling saves at least as much as they do. The best
// filling is a maximum-weight bipartite matching (blockers against
// non-tramplers and slots), solved exactly by maxWeightMatching.
//
// Where the relaxation is loose — a small blocker in a second slot
// behind one that already holds the trampler in full is credited with
// damage that was never coming — the bound only falls, never rises,
// and unblockedPower takes the larger of it and its own matching. It
// only asks when a trampler is in the swing; without one there are no
// slots and the answer is the matching's.
func trampleBound(st *state, defender string, attackers, blockers []*protocol.CardView) int {
	blockable := func(a, b *protocol.CardView) bool { return couldBlock(st, defender, a, b) }
	// profit[bi] is what blocker bi saves in each target: first the
	// non-tramplers, then every trampler's slots. Zero is no edge.
	profit := make([][]int, len(blockers))
	targets := 0
	addTarget := func(worth func(bi int) int) {
		for bi := range blockers {
			profit[bi] = append(profit[bi], worth(bi))
		}
		targets++
	}
	var tramplers []*protocol.CardView
	total := 0
	for _, a := range attackers {
		if a.Power <= 0 {
			continue
		}
		total += a.Power
		if hasKeyword(a, "trample") {
			tramplers = append(tramplers, a)
			continue
		}
		addTarget(func(bi int) int {
			if blockable(a, blockers[bi]) {
				return a.Power
			}
			return 0
		})
	}
	for _, t := range tramplers {
		credit := make([]int, len(blockers))
		least, able := 0, 0
		for bi, b := range blockers {
			if !blockable(t, b) {
				continue
			}
			credit[bi] = min(t.Power, trampleAbsorb(t, b))
			if able == 0 || credit[bi] < least {
				least = credit[bi]
			}
			able++
		}
		if able == 0 {
			continue
		}
		// Past slot ceil(p/least) no blocker's slotValue is positive,
		// and there are never more useful slots than blockers.
		slots := min((t.Power+least-1)/least, able)
		for j := 1; j <= slots; j++ {
			addTarget(func(bi int) int {
				c := credit[bi]
				return max(min(c, t.Power-(j-1)*c), 0)
			})
		}
	}
	return max(total-maxWeightMatching(profit, targets), 0)
}

// maxWeightMatching is the most total profit a matching of blockers
// (rows) to targets (columns) can collect, each blocker and each
// target used at most once; profit[b][t] ≤ 0 is no edge.
//
// Min-cost flow by successive shortest paths (SPFA), profits as
// negative costs: source → each blocker (capacity 1), blocker → target
// (1, −profit), target → sink (1). Each augmentation matches one more
// blocker, so there are at most len(profit) of them, and it stops at
// the first path that no longer pays. Boards are a few dozen creatures
// a side and a trampler has a handful of slots.
func maxWeightMatching(profit [][]int, targets int) int {
	nb := len(profit)
	src, sink := 0, nb+targets+1
	type edge struct{ to, rev, cap, cost int }
	g := make([][]edge, nb+targets+2)
	add := func(u, v, cost int) {
		g[u] = append(g[u], edge{v, len(g[v]), 1, cost})
		g[v] = append(g[v], edge{u, len(g[u]) - 1, 0, -cost})
	}
	for bi, row := range profit {
		add(src, 1+bi, 0)
		for ti, w := range row {
			if w > 0 {
				add(1+bi, 1+nb+ti, -w)
			}
		}
	}
	for ti := 0; ti < targets; ti++ {
		add(1+nb+ti, sink, 0)
	}

	const inf = 1 << 60
	n := len(g)
	dist := make([]int, n)
	inQ := make([]bool, n)
	prevNode, prevEdge := make([]int, n), make([]int, n)
	total := 0
	for {
		for i := range dist {
			dist[i], inQ[i] = inf, false
		}
		dist[src] = 0
		queue := []int{src}
		inQ[src] = true
		for len(queue) > 0 {
			u := queue[0]
			queue = queue[1:]
			inQ[u] = false
			for ei, e := range g[u] {
				if e.cap > 0 && dist[u]+e.cost < dist[e.to] {
					dist[e.to] = dist[u] + e.cost
					prevNode[e.to], prevEdge[e.to] = u, ei
					if !inQ[e.to] {
						inQ[e.to] = true
						queue = append(queue, e.to)
					}
				}
			}
		}
		if dist[sink] >= 0 {
			return total
		}
		for v := sink; v != src; v = prevNode[v] {
			e := &g[prevNode[v]][prevEdge[v]]
			e.cap--
			g[v][e.rev].cap++
		}
		total -= dist[sink]
	}
}
