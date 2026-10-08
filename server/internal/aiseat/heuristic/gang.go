package heuristic

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// gang.go is combat between one attacker and every creature blocking
// it (#1548): the gang block.
//
// A declare_blocker move is one pair, so a defender that wants two
// Ogres on a Wurm declares them one at a time, and the second has to be
// scored with the first already there. Before #1548 it was not: the
// block planner credited a kill only when the blocker it added killed
// alone, and the two-turn race predicted the same defender. A second
// Ogre on a Wurm one Ogre is already chumping (4+4 ≥ 7) was never worth
// declaring, even though the Wurm's 7 damage kills one Ogre either way
// and the second costs nothing.
//
// Three questions, all about one attacker and the whole group in front
// of it, and both sides of the table ask them:
//
//   - gangKills: does the group kill the attacker?
//   - connects: what does the attacker still put through to the player?
//   - assignable (with state.losses): how much damage does the attacker
//     get to divide among the group, and so which of them die?

// gangKills reports whether these blockers, all blocking atk, kill it
// when atk's controller assigns its damage to survive if it can.
//
// Combat damage comes in up to two steps (CR 510.4): creatures with
// first strike or double strike deal theirs first, and the rest — with
// every double striker a second time (CR 702.4b) — deal theirs second,
// but only if they are still on the battlefield. So an attacker with
// first strike may kill blockers before they deal regular damage, and
// which ones is its controller's choice: the engine lets the attacking
// player put the blockers in any order before assigning lethal damage
// down it (CR 510.1c), so any set of them whose lethal damage adds up
// to no more than its power can be killed first. The attacker survives
// when some such set leaves the damage it takes below its toughness
// with no deathtouch among what is left (CR 702.2b). Finding that set
// is a small knapsack — kill every deathtoucher, then remove as much
// regular damage as the rest of the power can buy — solved exactly.
//
// The rest is read as kills reads it: damage a blocker deals an
// attacker protected from it is prevented (CR 702.16e), marked damage
// counts, and an indestructible attacker never dies.
func gangKills(atk *protocol.CardView, blockers []*protocol.CardView) bool {
	if hasKeyword(atk, "indestructible") {
		return false
	}
	need := effectiveToughness(atk)
	// first is the damage atk takes in the first-strike step, from
	// blockers it cannot stop in time; second lists every blocker that
	// would deal it damage in the regular step.
	type hit struct {
		damage     int
		deathtouch bool
		// cost is what atk must assign this blocker to destroy it
		// first; -1 when it cannot.
		cost int
	}
	first, dealt := 0, false
	var second []hit
	for _, b := range blockers {
		if b.Power <= 0 || protectedFrom(atk, b) {
			continue
		}
		dealt = true
		dt := hasKeyword(b, "deathtouch")
		if firstStrikes(b) {
			if dt {
				return true // dies in the first step, whatever it does
			}
			first += b.Power
			if !hasKeyword(b, "double strike") {
				continue
			}
		}
		h := hit{damage: b.Power, deathtouch: dt, cost: -1}
		if firstStrikes(atk) {
			h.cost = lethalFrom(atk, b)
		}
		second = append(second, h)
	}
	if !dealt {
		return false
	}
	if first >= need {
		return true
	}
	// Kill every deathtoucher first — any one left alive is lethal —
	// and spend what is left of the power on the most regular damage
	// it can remove (0/1 knapsack over cost).
	budget := atk.Power
	total := first
	var items []hit
	for _, h := range second {
		if h.deathtouch {
			if h.cost < 0 || h.cost > budget {
				return true
			}
			budget -= h.cost
			continue
		}
		total += h.damage
		if h.cost >= 0 && h.cost <= budget {
			items = append(items, h)
		}
	}
	if total < need {
		return false
	}
	if len(items) == 0 || budget <= 0 {
		return true
	}
	best := make([]int, budget+1) // best[c]: most damage removed spending at most c
	for _, h := range items {
		for c := budget; c >= h.cost; c-- {
			if v := best[c-h.cost] + h.damage; v > best[c] {
				best[c] = v
			}
		}
	}
	return total-best[budget] >= need
}

// lethalFrom is the damage atk must assign blocker b to destroy it: 1
// from a deathtouch source (CR 702.2c), else its toughness less the
// damage already marked on it, never less than 1. -1 when atk cannot
// destroy it at all — no power, an indestructible blocker, or one
// protected from atk.
func lethalFrom(atk, b *protocol.CardView) int {
	if atk.Power <= 0 || hasKeyword(b, "indestructible") || protectedFrom(b, atk) {
		return -1
	}
	if hasKeyword(atk, "deathtouch") {
		return 1
	}
	return max(effectiveToughness(b), 1)
}

// diesFirst reports whether atk is destroyed in the first-strike damage
// step by these blockers, before any regular damage is dealt.
func diesFirst(atk *protocol.CardView, blockers []*protocol.CardView) bool {
	if hasKeyword(atk, "indestructible") {
		return false
	}
	first := 0
	for _, b := range blockers {
		if b.Power <= 0 || !firstStrikes(b) || protectedFrom(atk, b) {
			continue
		}
		if hasKeyword(b, "deathtouch") {
			return true
		}
		first += b.Power
	}
	return first > 0 && first >= effectiveToughness(atk)
}

// assignable is all the combat damage atk divides among these blockers:
// its power, twice over with double strike (CR 702.4b; damage marked
// in the first step still counts in the second) — unless the blockers'
// first-strike damage destroys it first. An attacker without first
// strike destroyed then deals nothing (CR 510.4, #1549); one with it
// deals its first-strike damage at the same time, and no more.
func assignable(atk *protocol.CardView, blockers []*protocol.CardView) int {
	if atk.Power <= 0 {
		return 0
	}
	if diesFirst(atk, blockers) {
		if firstStrikes(atk) {
			return atk.Power
		}
		return 0
	}
	return combatDamage(atk)
}

// orderedKills is the blockers atk's damage kills when it is assigned
// the way this bot assigns it: legal's canonical assignment, the one
// split the enumerator offers, which kills the set of blockers worth
// most that its damage can buy (#2692). CR 510.1c lets the attacking
// player divide the damage among the blockers as they choose, so any
// set whose lethal damage fits in the power can die. A kill is worth
// one plus the blocker's power and toughness, the price legal puts on
// it; a blocker the damage cannot destroy is worth nothing. Both pick
// the set with legal.KillingSet, so the bot plans the combat it is
// offered. It is the
// attacker's side of losses: losses prices a block for the defender,
// who has to assume the worst; this is what the bot's own swing will
// actually kill.
func orderedKills(atk *protocol.CardView, blockers []*protocol.CardView) []*protocol.CardView {
	n := len(blockers)
	cost := make([]int, n)
	value := make([]int, n)
	for i, b := range blockers {
		cost[i] = 1
		if !hasKeyword(atk, "deathtouch") {
			cost[i] = max(effectiveToughness(b), 1)
		}
		if lethalFrom(atk, b) >= 0 {
			value[i] = 1 + max(b.Power, 0) + max(b.Toughness, 0)
		}
	}
	kill := legal.KillingSet(cost, value, assignable(atk, blockers))
	var out []*protocol.CardView
	for i, b := range blockers {
		if kill[i] {
			out = append(out, b)
		}
	}
	return out
}

// connects is what attacker atk puts through to the player it attacks
// with these blockers in front of it: all of its power unblocked,
// nothing once blocked — unless it tramples, when whatever its blockers
// do not absorb goes over (CR 702.19b, overflowPast). So a second
// blocker on a chumped Wurm saves damage too: a Bear in front of a 7/7
// lets 5 through, a Bear and an Ogre let 1.
func connects(atk *protocol.CardView, blockers []*protocol.CardView) int {
	left := max(atk.Power, 0)
	if len(blockers) == 0 {
		return left
	}
	if !hasKeyword(atk, "trample") {
		return 0
	}
	for _, b := range blockers {
		left = overflowPast(atk, b, left)
	}
	return left
}

// blockersOn is every creature on the battlefield blocking attacker
// atkID, the ones blocking several attackers included (#1706).
func blockersOn(st *state, atkID string) []*protocol.CardView {
	var out []*protocol.CardView
	for i := range st.view.Battlefield.Cards {
		c := &st.view.Battlefield.Cards[i]
		if c.BlockingTarget == atkID {
			out = append(out, c)
			continue
		}
		for _, id := range c.BlockingTargets {
			if id == atkID {
				out = append(out, c)
				break
			}
		}
	}
	return out
}

// lossValue is what the defender loses when these blockers take atk:
// the creatures its damage kills (state.losses), a deathtoucher's
// victims priced a quarter higher — it eats whatever blocks it, so the
// best creature on the board should not be the one fed to it.
func (st *state) lossValue(atk *protocol.CardView, blockers []*protocol.CardView) float64 {
	v := 0.0
	for _, blk := range st.losses(atk, blockers) {
		v += st.w.CombatValue(blk)
		if hasKeyword(atk, "deathtouch") {
			v += st.w.CombatValue(blk) * 0.25
		}
	}
	return v
}

// gangJoin is the cheapest way for the defender to turn a block that
// does not kill attacker a into one that does (#1548): the blockers out
// of cands to add to have, by index. The cost of a join is what the
// group loses with it over what have already loses (lossValue). ok is
// false when no join kills a, or when the cheapest one costs at least a
// is worth — a gang that trades two Ogres for a Bear is not one anyone
// makes.
//
// It tries every single blocker and every pair, then cands biggest
// first until the group kills, and keeps the cheapest, fewest blockers
// on a tie. A board's worth of identical Ogres makes that a handful of
// distinct tries, and nothing larger than a pair is searched for
// exactly: a gang of four is rare, and the biggest-first fallback still
// finds one.
func (st *state) gangJoin(a *protocol.CardView, have, cands []*protocol.CardView) (join []int, ok bool) {
	if len(cands) == 0 || !mightKill(a, have, cands) || gangKills(a, have) {
		return nil, false
	}
	cost := 0.0
	base := st.lossValue(a, have)
	gang := make([]*protocol.CardView, len(have), len(have)+len(cands))
	copy(gang, have)
	try := func(idx []int) {
		g := gang[:len(have)]
		for _, i := range idx {
			g = append(g, cands[i])
		}
		if !gangKills(a, g) {
			return
		}
		c := st.lossValue(a, g) - base
		if !ok || c < cost || (c == cost && len(idx) < len(join)) {
			join, cost, ok = append([]int(nil), idx...), c, true
		}
	}
	// rep[i] is the first candidate interchangeable with cands[i], so a
	// board of identical Ogres is one single and one pair to try.
	rep := make([]int, len(cands))
	for i := range cands {
		rep[i] = i
		for j := 0; j < i; j++ {
			if rep[j] == j && sameBlocker(cands[i], cands[j]) {
				rep[i] = j
				break
			}
		}
	}
	for i := range cands {
		if rep[i] == i {
			try([]int{i})
		}
	}
	seenPair := map[[2]int]bool{}
	for i := range cands {
		for j := i + 1; j < len(cands); j++ {
			key := [2]int{min(rep[i], rep[j]), max(rep[i], rep[j])}
			if seenPair[key] {
				continue
			}
			seenPair[key] = true
			try([]int{i, j})
		}
	}
	if !ok {
		order := byPower(cands)
		for k := 3; k <= len(order); k++ {
			idx := append([]int(nil), order[:k]...)
			try(idx)
			if ok {
				break
			}
		}
	}
	if !ok || cost >= st.w.CombatValue(a) {
		return nil, false
	}
	return join, true
}

// sameBlocker reports whether two creatures are the same choice to
// gangJoin: equal in everything gangKills, losses and CombatValue read.
func sameBlocker(a, b *protocol.CardView) bool {
	return a.Power == b.Power && a.Toughness == b.Toughness && a.DamageMarked == b.DamageMarked &&
		a.TypeLine == b.TypeLine && slices.Equal(a.Abilities, b.Abilities) &&
		slices.Equal(a.Colors, b.Colors) && slices.Equal(a.Protection, b.Protection)
}

// mightKill is gangJoin's quick exit: whether have and every candidate
// together could deal atk lethal damage at all, before any timing.
func mightKill(atk *protocol.CardView, have, cands []*protocol.CardView) bool {
	if hasKeyword(atk, "indestructible") {
		return false
	}
	total := 0
	for _, group := range [][]*protocol.CardView{have, cands} {
		for _, b := range group {
			if b.Power <= 0 || protectedFrom(atk, b) {
				continue
			}
			if hasKeyword(b, "deathtouch") {
				return true
			}
			total += combatDamage(b)
		}
	}
	return total >= effectiveToughness(atk)
}
