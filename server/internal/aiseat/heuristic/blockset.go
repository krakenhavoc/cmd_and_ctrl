package heuristic

import (
	"slices"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// blockset.go is #3066: the defender's blocks checked as a whole set
// against lethal damage (Config.BlockSetSurvives).
//
// decideBlock declares one block at a time and scores each on its own,
// so a defender on 3 facing a Bear 2/2 and a 3/3 commander, with its
// own commander as its only blocker, took the Bear block (+6.90: it
// kills the Bear and saves 2) and died to the commander's 3. The block
// that keeps it alive is commander on commander, taking 2.
//
// The check runs only when finishing the declaration now would kill
// the bot. Blocking more never lets more damage through, so a set that
// survives as it stands survives every block added to it, and the
// check has nothing to say. When it would die, each block on offer is
// asked whether some way of spending the creatures still free after it
// survives (survivable). If one does, decideBlock chooses among those
// blocks only, by its usual score, even a score at or below zero:
// finishing would kill. If none does, decideBlock is unchanged.

// poisonLethal is CR 104.3d: ten or more poison counters lose.
const poisonLethal = 10

// blockSearchBudget caps the search over the free creatures' blocks,
// in boards visited, for one candidate block. A search that runs out
// answers "unknown", and an unknown block is neither preferred nor
// ruled out.
const blockSearchBudget = 2000

// blockSet is the incoming combat as the defender can still change it:
// every attacker pointed at the bot, what already blocks each, and the
// bot's creatures still free to block, with which attackers each may
// block.
type blockSet struct {
	atks []*protocol.CardView
	// on is the blockers in front of each attacker: the declared ones,
	// plus whatever a candidate or the search adds.
	on [][]*protocol.CardView
	// free is the bot's untapped creatures not yet blocking, and
	// can[i] the attackers free[i] may block, as indices into atks.
	free []*protocol.CardView
	can  [][]int

	life, poison int
	cantLoseLife bool
	// cmdrTaken is the combat damage each commander has dealt the bot
	// so far, and cmdrLethal the total from one that loses (CR
	// 903.10a; the table's setting, 21 by default).
	cmdrTaken  map[string]int
	cmdrLethal int

	nodes     int
	exhausted bool
}

// newBlockSet reads the incoming combat off the view. Which free
// creature may block which attacker is read off the moves on offer —
// the enumerator's exact CR 509.1 answer for this window — so a pair it
// does not offer is never searched.
func newBlockSet(st *state, moves []legal.Move) *blockSet {
	b := &blockSet{life: StartingLife, cmdrLethal: 21}
	if st.myEval != nil {
		b.life = st.myEval.Life
		b.cantLoseLife = st.myEval.CantLoseLife
	}
	if st.seat != nil {
		b.poison = st.seat.Poison
		b.cmdrTaken = st.seat.CommanderDamage
	}
	if s := st.view.Settings; s != nil && s.CommanderDamage > 0 {
		b.cmdrLethal = s.CommanderDamage
	}
	idx := map[string]int{}
	for i := range st.view.Battlefield.Cards {
		c := &st.view.Battlefield.Cards[i]
		if c.AttackingTarget == st.me {
			idx[c.InstanceID] = len(b.atks)
			b.atks = append(b.atks, c)
			b.on = append(b.on, blockersOn(st, c.InstanceID))
		}
	}
	// Who may block whom, from every pair any block move names.
	pairs := map[string][]int{}
	for i := range moves {
		if moves[i].Kind != legal.KindBlock {
			continue
		}
		atk, blks := st.blockMove(moves[i])
		if atk == nil {
			continue
		}
		a, ok := idx[atk.InstanceID]
		if !ok {
			continue
		}
		for _, blk := range blks {
			if !slices.Contains(pairs[blk.InstanceID], a) {
				pairs[blk.InstanceID] = append(pairs[blk.InstanceID], a)
			}
		}
	}
	for i := range st.view.Battlefield.Cards {
		c := &st.view.Battlefield.Cards[i]
		can, ok := pairs[c.InstanceID]
		if !ok || c.BlockingTarget != "" || len(c.BlockingTargets) > 0 {
			continue
		}
		slices.Sort(can)
		b.free = append(b.free, c)
		b.can = append(b.can, can)
	}
	return b
}

// through is what attacker i deals the bot with its blockers as they
// stand: connects, with a menace attacker that has fewer than two
// blockers counted unblocked (CR 702.110b; the search may put one
// creature in front of it on the way to two), and a double striker's
// damage dealt twice (CR 702.4b).
func (b *blockSet) through(i int) int {
	atk, on := b.atks[i], b.on[i]
	if hasKeyword(atk, "menace") && len(on) < 2 {
		on = nil
	}
	d := connects(atk, on)
	if hasKeyword(atk, "double strike") {
		d *= 2
	}
	return d
}

// dies reports whether the blocks as they stand kill the bot: its life
// (unless it can't lose for life), ten poison counters from infect and
// toxic (CR 702.90b, 702.164c), or 21 combat damage from one commander.
func (b *blockSet) dies() bool {
	lost, poison := 0, b.poison
	for i, atk := range b.atks {
		d := b.through(i)
		if d <= 0 {
			continue
		}
		if hasKeyword(atk, "infect") {
			poison += d
		} else {
			lost += d
		}
		poison += toxic(atk)
		if atk.IsCommander && b.cmdrTaken[atk.InstanceID]+d >= b.cmdrLethal {
			return true
		}
	}
	return (!b.cantLoseLife && b.life-lost <= 0) || poison >= poisonLethal
}

// toxic is the total of a creature's toxic instances (CR 702.164c:
// each one triggers separately).
func toxic(c *protocol.CardView) int {
	n := 0
	for _, a := range c.Abilities {
		word, num, ok := strings.Cut(strings.ToLower(strings.TrimSpace(a)), " ")
		if !ok || word != "toxic" {
			continue
		}
		v := 0
		for _, r := range num {
			if r < '0' || r > '9' {
				v = 0
				break
			}
			v = v*10 + int(r-'0')
		}
		n += v
	}
	return n
}

// hopeless is a cheap bound on the free creatures from k on: each can
// stop at most one attacker's damage, so the bot takes at least
// everything less the k-th-onward creatures' best attackers. A
// commander's lethal hit that none of them may block kills outright.
// True means no way of spending them survives.
func (b *blockSet) hopeless(k int) bool {
	blockable := make([]bool, len(b.atks))
	for _, can := range b.can[k:] {
		for _, a := range can {
			blockable[a] = true
		}
	}
	var lifeHits, poisonHits []int
	lost, poison := 0, b.poison
	for i, atk := range b.atks {
		d := b.through(i)
		if d <= 0 {
			continue
		}
		life, pois := d, toxic(atk)
		if hasKeyword(atk, "infect") {
			life, pois = 0, pois+d
		}
		lost += life
		poison += pois
		if blockable[i] {
			lifeHits = append(lifeHits, life)
			poisonHits = append(poisonHits, pois)
		} else if atk.IsCommander && b.cmdrTaken[atk.InstanceID]+d >= b.cmdrLethal {
			return true
		}
	}
	n := len(b.free) - k
	if !b.cantLoseLife && b.life-(lost-topSum(lifeHits, n)) <= 0 {
		return true
	}
	return poison-topSum(poisonHits, n) >= poisonLethal
}

// topSum is the sum of the n largest of xs.
func topSum(xs []int, n int) int {
	slices.SortFunc(xs, func(a, b int) int { return b - a })
	s := 0
	for i := 0; i < n && i < len(xs); i++ {
		s += xs[i]
	}
	return s
}

// survivable reports whether some way of spending the free creatures
// from k on leaves the bot alive. Each creature blocks one attacker
// that still gets damage through, or blocks nothing. It stops at the
// first set that survives, and gives up (exhausted) past the budget.
func (b *blockSet) survivable(k int) bool {
	b.nodes++
	if b.nodes > blockSearchBudget {
		b.exhausted = true
		return false
	}
	if !b.dies() {
		return true
	}
	if k >= len(b.free) || b.hopeless(k) {
		return false
	}
	blk := b.free[k]
	for _, a := range b.can[k] {
		if b.through(a) <= 0 {
			continue
		}
		b.on[a] = append(b.on[a], blk)
		ok := b.survivable(k + 1)
		b.on[a] = b.on[a][:len(b.on[a])-1]
		if ok || b.exhausted {
			return ok
		}
	}
	return b.survivable(k + 1)
}

// blockOutcome is what survivable says about one block move.
type blockOutcome int

const (
	blockUnknown blockOutcome = iota
	blockSurvives
	blockDies
)

// after asks whether the bot can still survive once this block move is
// declared: its creatures go in front of the attacker and leave the
// free list.
func (b *blockSet) after(atk *protocol.CardView, adding []*protocol.CardView) blockOutcome {
	a := slices.IndexFunc(b.atks, func(c *protocol.CardView) bool { return c.InstanceID == atk.InstanceID })
	if a < 0 {
		return blockUnknown
	}
	free, can := b.free, b.can
	defer func() { b.free, b.can = free, can }()
	b.free, b.can = nil, nil
	for i, c := range free {
		if !slices.ContainsFunc(adding, func(x *protocol.CardView) bool { return x.InstanceID == c.InstanceID }) {
			b.free = append(b.free, c)
			b.can = append(b.can, can[i])
		}
	}
	on := b.on[a]
	b.on[a] = append(append(make([]*protocol.CardView, 0, len(on)+len(adding)), on...), adding...)
	defer func() { b.on[a] = on }()

	b.nodes, b.exhausted = 0, false
	ok := b.survivable(0)
	switch {
	case ok:
		return blockSurvives
	case b.exhausted:
		return blockUnknown
	}
	return blockDies
}
