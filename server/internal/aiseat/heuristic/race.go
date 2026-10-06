package heuristic

import (
	"fmt"
	"sort"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// race.go is the two-turn race (#1409): the attack decision's look one
// turn past this one.
//
// lethalPush asks "does everything I have kill them now?", and the
// per-creature attackValue asks "is this one attack a good trade?".
// Between the two is the board that stalled the S31 heuristic gate: two
// seats at low life behind full boards, one side a Drake up. No all-in
// is lethal this turn, every single attack is blocked at a loss, so
// neither seat ever declares the first attacker and the game waits for
// a library to run out. A player sees it at once: swing the Drakes,
// they have to trade or take it, and the Drake that is left over
// finishes them next turn — while nothing they have left can kill me
// on the turn in between.
//
// That is the whole rule, in three numbers, all computed with the same
// blocker/attacker matching lethalPush uses (unblockedPower):
//
//  1. NOW: the damage this swing connects for after the defender
//     blocks to survive, losing as few creatures as it can.
//  2. NEXT: what my survivors connect for next turn into what the
//     defender has left, assuming every one of its survivors stays
//     home to block.
//  3. CRACK-BACK: what every opponent's creatures connect for on the
//     turn in between, attacking into the creatures I kept home,
//     assuming every one of them attacks me.
//
// The race commits when NOW + NEXT is at least the defender's life and
// CRACK-BACK is less than mine. The last condition is what makes it
// never suicidal, and it is pessimistic on purpose on every axis the
// estimate has: none of my blocked attackers is home to block, the
// defender keeps every creature it did not have to lose, it is assumed
// to both attack with all of them AND keep all of them home, every
// opponent swings at me rather than at each other, a trampler is only
// held by a blocker that can absorb all of it, and a menace creature
// counts as unblockable. An estimate that is wrong errs toward not
// racing, which leaves the bot exactly where it was before #1409.
//
// NOW and NEXT count trample the other way round (#1504): a blocked
// trampler connects for whatever its blockers do not absorb, so a Bear
// in front of my Wurm still lets 5 through. Without it a one-Wurm edge
// was invisible, where a one-Drake edge was not. Where that count has
// to guess at the defender's blocks it takes a lower bound
// (trampleBound), so it too errs toward not racing.
//
// NEXT also keeps a blocked attacker of mine that lives (#1527). It
// used to count every one of them dead, so the Wurm a Bear chumped this
// turn was gone by the next, and a trampler's edge was cashed only when
// this turn's overflow closed the gap alone. A blocked attacker now
// lives into NEXT when the blockers the defender's model put in front
// of it, joined by every blocker it left spare, cannot kill it
// (survives). It is still never home: the crack-back check does not
// change.
//
// The defender gang-blocks to kill (#1548), as decideBlock does: a
// second Ogre on a Wurm one Ogre is chumping kills it for nothing. So
// the model adds spare blockers to a block that does not kill when the
// kill is worth what they cost (blockToSurvive), they soak up a
// trampler's overflow as they go, and survives counts the rest of the
// spare pool as joining too. That takes away a lot of two-turn kills —
// the full seed-1409 Wurm mirror has none left — and attrition.go is
// what cashes those edges over more turns.
//
// The defender has a second answer, and the race has to beat it too
// (#2310): block with everything whose block costs it nothing — a
// creature that lives, or its commander, since CR 903.9a sends a dying
// commander to the command zone and it is back before my next attack.
// Damage that answer stops is not damage the race may bank. Counting
// it ran a lone commander into the defender's commander for a dozen
// turns: both were cast again every turn, and the same race came round
// each time (raceNumbers).
//
// The swing it sends is the SMALLEST that wins: attackers are tried
// evasive-first (fewest possible blockers, then biggest), and the plan
// is the shortest prefix of that order that races. Everything outside
// it stays home, because the reserve is what the crack-back check
// counted on.
//
// A perfect mirror has no race — equal boards trade evenly and nothing
// is left over for NEXT — and that is correct: the race converts an
// edge, it does not invent one.

// racePlan is a committed two-turn race against one defender — or an
// attrition plan (attrition.go), whose swing is sent the same way.
type racePlan struct {
	target string
	// swing is every attacker the plan sends at target, in the order
	// it should declare them — already-declared ones included.
	swing []*protocol.CardView
	// now, next and crack are the three numbers the plan was accepted
	// on, kept for the decision's Reason.
	now, next, crack int
	life, myLife     int
	// reason, when set, is the decision's Reason in place of the
	// two-turn race's.
	reason string
}

// planRace looks for a two-turn race against any opponent, focus first.
// Nil when there is none, which is every board without an edge.
func (p *Policy) planRace(st *state, moves []legal.Move, focus string) *racePlan {
	if st.myEval == nil || st.myEval.Life <= 0 {
		return nil
	}
	for _, def := range focusFirst(st.opps, focus) {
		if def.Life <= 0 {
			continue
		}
		committed, joinable := raceSwing(st, moves, def)
		if len(joinable) == 0 {
			continue
		}
		if plan := p.raceWith(st, def, committed, joinable, nil); plan != nil {
			return plan
		}
	}
	return nil
}

// focusFirst is the opponents in threat order with focus moved to the
// front.
func focusFirst(opps []*SeatEval, focus string) []*SeatEval {
	order := make([]*SeatEval, 0, len(opps))
	for _, o := range opps {
		if o.ID == focus {
			order = append([]*SeatEval{o}, order...)
			continue
		}
		order = append(order, o)
	}
	return order
}

// raceSwing is what a swing at def is made of: the creatures already
// committed to it, part of every swing, and the ones still able to
// join, in the order the race tries them.
func raceSwing(st *state, moves []legal.Move, def *SeatEval) (committed, joinable []*protocol.CardView) {
	for i := range st.view.Battlefield.Cards {
		c := &st.view.Battlefield.Cards[i]
		if c.Controller == st.me && isCreature(c) && c.AttackingTarget == def.ID {
			committed = append(committed, c)
		}
	}
	// Still able to join, read off the move list so every restriction
	// and every unpayable attack tax the enumerator applied is honoured.
	seen := map[string]bool{}
	for i := range moves {
		if moves[i].Kind != legal.KindAttack {
			continue
		}
		ap := decode[attackParams](moves[i].Params)
		if ap.Target != def.ID || seen[ap.Attacker] {
			continue
		}
		seen[ap.Attacker] = true
		if c := st.bf[ap.Attacker]; c != nil {
			joinable = append(joinable, c)
		}
	}
	evasiveFirst(st, def, joinable)
	return committed, joinable
}

// evasiveFirst sorts would-be attackers in the order the race tries
// them: fewest possible blockers first, then biggest. Stable, so board
// order breaks ties and the plan never depends on a sort's whim.
func evasiveFirst(st *state, def *SeatEval, joinable []*protocol.CardView) {
	blockers := defenderBlockers(st, def.ID)
	answers := make(map[string]int, len(joinable))
	for _, a := range joinable {
		for _, b := range blockers {
			if couldBlock(st, def.ID, a, b) {
				answers[a.InstanceID]++
			}
		}
	}
	sort.SliceStable(joinable, func(i, j int) bool {
		ai, aj := answers[joinable[i].InstanceID], answers[joinable[j].InstanceID]
		if ai != aj {
			return ai < aj
		}
		return joinable[i].Power > joinable[j].Power
	})
}

// raceWith tries the prefixes of the evasion-sorted attack order
// against one defender and returns the first that races. budget, when
// not nil, is the attrition search's (attrition.go): every swing tried
// spends one, and an empty budget finds nothing.
func (p *Policy) raceWith(st *state, def *SeatEval, committed, joinable []*protocol.CardView, budget *int) *racePlan {
	blockers := defenderBlockers(st, def.ID)
	// The empty prefix is tried when something is already declared: a
	// swing that already races needs nobody else, and adding to it only
	// spends the reserve.
	first := 1
	if len(committed) > 0 {
		first = 0
	}
	for k := first; k <= len(joinable); k++ {
		if budget != nil {
			if *budget <= 0 {
				return nil
			}
			*budget--
		}
		swing := make([]*protocol.CardView, 0, len(committed)+k)
		swing = append(swing, committed...)
		swing = append(swing, joinable[:k]...)
		e := p.raceNumbers(st, def, swing, blockers)
		if e.now+e.next >= def.Life && e.crack < st.myEval.Life {
			return &racePlan{
				target: def.ID, swing: swing,
				now: e.now, next: e.next, crack: e.crack,
				life: def.Life, myLife: st.myEval.Life,
			}
		}
	}
	return nil
}

// raceEval is the race's estimate for one swing at one defender: its
// three numbers, the defender's blocks, and which of my attackers those
// blocks kill.
type raceEval struct {
	// now and next are the pair from whichever of the defender's two
	// answers leaves it more life after next turn (#2310); crack is the
	// larger of the two answers' crack-backs.
	now, next, crack int
	// d and mineDead are blockToSurvive's answer: its blocks, and every
	// attacker of mine it kills — blocked, and not surviving the blocks
	// (defence.survives). The attrition search plays them forward.
	d        defence
	mineDead map[string]bool
	// free is what the swing connects for past the defender's free
	// blocks (freeAnswer).
	free int
}

// answer is one way the defender can meet a swing, as the race counts
// it: the damage that connects, every attacker of mine it blocks and
// the ones it kills, the defender's creatures that die, and the dead
// commanders it casts again (recast). A recast commander is back to
// block my next attack, but it was cast on the defender's own turn and
// is too new to attack on it (CR 302.6), so it counts for NEXT and not
// for the crack-back.
type answer struct {
	through  int
	blocked  map[string]bool
	mineDead map[string]bool
	dead     map[string]bool
	recast   map[string]bool
}

// raceNumbers is the three-number estimate for one swing at def.
//
// The defender has two cheap answers, and the race has to win against
// both (#2310). blockToSurvive's takes whatever damage it survives, and
// blocks only to live or to kill for free. The free answer (freeAnswer)
// takes as little as it can without losing anything for good: it
// blocks with every creature whose block costs it nothing (freeBlock),
// a blocker that lives, or its commander. Without the second answer
// the full seed-1409 Wurm mirror looped for a dozen turns: a lone 3/3
// commander raced for "3 now + 4 next", the defender blocked it with
// its own commander, both were cast again, and the same race came
// round. Against the free answer that swing is 0 now and 2 next, which
// is no race.
//
// So NOW and NEXT are the pair from whichever answer leaves the
// defender more life after next turn — under the free answer an
// attacker of mine a free block kills is dead for NEXT, and the
// commander is back to block — and CRACK-BACK is the larger of the
// two: the free answer loses nothing that can attack, so whatever
// blockToSurvive's chumps took out of the crack-back is back in it.
// The free answer is only an answer when the defender survives it; if
// it still dies to the swing it has to block the first way.
func (p *Policy) raceNumbers(st *state, def *SeatEval, swing, blockers []*protocol.CardView) raceEval {
	d := p.blockToSurvive(st, def, swing, blockers)
	took := answer{through: d.through, blocked: d.blockedMine, mineDead: map[string]bool{}, dead: d.deadDef}
	for _, a := range swing {
		// Counted dead unless no block the defender could make
		// kills it (#1527): a Wurm chumped by the last Bear
		// tramples over again next turn.
		if d.blockedMine[a.InstanceID] && !d.survives(st, def.ID, a) {
			took.mineDead[a.InstanceID] = true
		}
	}
	e := raceEval{now: took.through, d: d, mineDead: took.mineDead}
	e.next, e.crack = st.raceAfter(def, swing, took)

	free := st.freeAnswer(def.ID, swing, blockers)
	e.free = free.through
	if free.through < def.Life {
		next, crack := st.raceAfter(def, swing, free)
		if free.through+next < e.now+e.next {
			e.now, e.next = free.through, next
		}
		e.crack = max(e.crack, crack)
	}
	return e
}

// raceAfter is NEXT and CRACK-BACK after the defender meets swing with
// ans.
func (st *state) raceAfter(def *SeatEval, swing []*protocol.CardView, ans answer) (next, crack int) {
	inSwing := make(map[string]bool, len(swing))
	for _, c := range swing {
		inSwing[c.InstanceID] = true
	}
	// Home: what I will have untapped on the turn in between. A swing
	// attacker stays untapped only with vigilance, and only if nothing
	// blocked it. A blocked attacker is never home, even one that lives
	// (#1527): the crack-back check stays exactly as pessimistic as it
	// was, and only NEXT hears about the survivors.
	var home, second []*protocol.CardView
	for i := range st.view.Battlefield.Cards {
		c := &st.view.Battlefield.Cards[i]
		if c.Controller != st.me || !isCreature(c) {
			continue
		}
		if ans.blocked[c.InstanceID] {
			if !ans.mineDead[c.InstanceID] {
				second = append(second, c)
			}
			continue
		}
		if c.AttackingTarget != "" && !inSwing[c.InstanceID] {
			continue // attacking somebody else, and may die doing it
		}
		if inSwing[c.InstanceID] {
			if hasKeyword(c, "vigilance") && !c.Tapped {
				home = append(home, c)
			}
		} else if !c.Tapped {
			home = append(home, c)
		}
		if !hasKeyword(c, "defender") {
			second = append(second, c)
		}
	}

	// CRACK-BACK: every opponent swings everything at me — except a
	// defender this swing already killed, and the defender's creatures
	// that died, a recast commander among them.
	for _, o := range st.opps {
		if o.ID == def.ID && ans.through >= def.Life {
			continue
		}
		var theirs []*protocol.CardView
		for i := range st.view.Battlefield.Cards {
			c := &st.view.Battlefield.Cards[i]
			if c.Controller != o.ID || !isCreature(c) || hasKeyword(c, "defender") {
				continue
			}
			if o.ID == def.ID && (ans.dead[c.InstanceID] || ans.recast[c.InstanceID]) {
				continue
			}
			theirs = append(theirs, c)
		}
		crack += crackBackPower(st, theirs, home)
	}

	// NEXT: my survivors into everything the defender kept, a recast
	// commander included.
	if ans.through >= def.Life {
		return 0, crack
	}
	var kept []*protocol.CardView
	for i := range st.view.Battlefield.Cards {
		c := &st.view.Battlefield.Cards[i]
		if c.Controller == def.ID && isCreature(c) && !ans.dead[c.InstanceID] {
			kept = append(kept, c)
		}
	}
	return unblockedPower(st, def.ID, second, kept), crack
}

// freeBlock reports whether blocking a with b costs the defender
// nothing it keeps (#2310): b lives through it, or b is its commander.
// CR 903.9a lets a dying commander go back to the command zone; the
// defender casts it again on its own turn, and it is back in time to
// block my next attack. It is the attrition horizon's definition of a
// free block (#1548), now the race's too.
//
// It does not ask whether the defender can pay the commander tax (CR
// 903.8) to cast it again. Measured on the heuristic gate, asking made
// the games longer, not shorter: a commander the tax had priced out
// let a smaller swing qualify, and the race and the attrition horizon
// send the smallest swing that does. Assuming the commander always
// comes back errs toward not racing, the race's usual direction.
func freeBlock(st *state, defender string, a, b *protocol.CardView) bool {
	return couldBlock(st, defender, a, b) && (b.IsCommander || !blockerDies(a, b))
}

// freeAnswer is the defender's free answer to a swing (#2310): block
// with every creature whose block is free (freeBlock), and take the
// rest. Its damage is unblockedPower's matching over the free blocks
// alone, and a blocked attacker counts as held in full, trample or
// not, which only ever lowers it — so it too errs toward not racing.
// An attacker of mine it blocks dies when the blocker, joined by every
// free blocker the matching left spare, kills it — the gang
// blockToSurvive's answer assumes too (defence.survives). Nothing of
// the defender's dies for good: a blocker that lives lives, and a
// commander that dies is cast again.
func (st *state) freeAnswer(defender string, swing, blockers []*protocol.CardView) answer {
	free := func(a, b *protocol.CardView) bool { return freeBlock(st, defender, a, b) }
	through, stoppedBy := matchBlocks(swing, blockers, free)
	ans := answer{through: through, blocked: map[string]bool{}, mineDead: map[string]bool{}, recast: map[string]bool{}}
	used := make([]bool, len(blockers))
	for _, bi := range stoppedBy {
		if bi >= 0 {
			used[bi] = true
		}
	}
	for ai, bi := range stoppedBy {
		if bi < 0 {
			continue
		}
		a, b := swing[ai], blockers[bi]
		ans.blocked[a.InstanceID] = true
		if blockerDies(a, b) && !hasKeyword(b, "haste") {
			// Cast again on the defender's turn, it cannot attack on
			// it — unless it has haste (CR 702.10).
			ans.recast[b.InstanceID] = true
		}
		gang := []*protocol.CardView{b}
		for si, s := range blockers {
			if !used[si] && free(a, s) {
				gang = append(gang, s)
			}
		}
		if gangKills(a, gang) {
			ans.mineDead[a.InstanceID] = true
		}
	}
	return ans
}

// defence is blockToSurvive's answer to one swing.
type defence struct {
	// through is the damage that connects.
	through int
	// deadDef is the defender's creatures that die blocking.
	deadDef map[string]bool
	// blockedMine is every attacker of mine that was blocked at all.
	blockedMine map[string]bool
	// blockedBy is, per blocked attacker, the blockers the model put in
	// front of it; spare is every blocker the model left unassigned.
	blockedBy map[string][]*protocol.CardView
	spare     []*protocol.CardView
	// guessed marks the fallback answer, where the model did not find
	// the defender's blocks and assumed every attacker was blocked by
	// something it cannot name.
	guessed bool
}

// survives reports whether my blocked attacker a lives through the
// defender's blocks (#1527). A blocked attacker used to be counted dead
// whatever blocked it, so a Wurm chumped by a Bear was as gone as one
// traded with a Wurm, and a trampler's edge was cashed only when this
// turn's overflow alone closed the gap.
//
// It says yes only when the blockers the model put in front of it,
// joined by every blocker the model left spare that could block it
// too, do not kill it (gangKills). The defender gang-blocks to kill
// (#1548): decideBlock values a second Ogre on a chumped Wurm as the
// kill it is, so the model assumes the defender makes it. A spare that
// kills alone is the same question — a group with it in kills whenever
// it does — so "blocks to kill when it can" needs no separate swap.
//
// Each attacker is asked against the whole spare pool, as if nothing
// else needed those blockers, so two chumped Wurms can both be counted
// dead to the same spare Ogres. That only ever says "dead" about an
// attacker that lives, which costs a race and never a suicidal swing:
// survivors are counted into NEXT, never into crack-back.
//
// Where the model only guessed at the blocks, nothing survives.
func (d *defence) survives(st *state, defender string, a *protocol.CardView) bool {
	if d.guessed {
		return false
	}
	gang := append([]*protocol.CardView(nil), d.blockedBy[a.InstanceID]...)
	for _, b := range d.spare {
		if couldBlock(st, defender, a, b) {
			gang = append(gang, b)
		}
	}
	return !gangKills(a, gang)
}

// blockerDies reports whether blocker b, blocking a alone, is killed
// by a (#1549). It is kills with the first-strike timing kills does
// not read: a blocker that deals first-strike damage when a does not
// (CR 510.4, CR 702.7b) and kills a with that damage first takes none
// back, so it lives. A double striker's first-strike hit is one power
// (CR 702.4b), not two. If a also has first strike or double strike,
// both deal damage in the same step and b dies as kills says.
func blockerDies(a, b *protocol.CardView) bool {
	if !kills(a, b) {
		return false
	}
	return !(firstStrikes(b) && !firstStrikes(a) && firstStrikeKills(b, a))
}

// firstStrikeKills reports whether b's damage in the first combat
// damage step alone kills a: one power, not a double striker's two.
func firstStrikeKills(b, a *protocol.CardView) bool {
	if hasKeyword(a, "indestructible") || b.Power <= 0 || protectedFrom(a, b) {
		return false
	}
	return hasKeyword(b, "deathtouch") || b.Power >= effectiveToughness(a)
}

// combatDamage is all the damage creature c deals in one combat: its
// power, twice over with double strike.
func combatDamage(c *protocol.CardView) int {
	if hasKeyword(c, "double strike") {
		return 2 * c.Power
	}
	return c.Power
}

// blockToSurvive is the defender's answer to a swing: first every
// block it survives (those cost it nothing), and only if the damage
// still getting through would kill it, the cheapest blocks it does not
// survive, biggest attacker first, until it would live. It returns the
// damage that connects, the defender's creatures that die blocking,
// every attacker of mine that was blocked at all and who blocked it,
// and the blockers it left spare. A blocked trampler connects for what
// its blockers do not absorb (#1504), so a chump in front of one saves
// only the chump's toughness. Last, it gang-blocks to kill (#1548):
// spare blockers join a block that does not kill its attacker when
// that turns it into one cheaply enough (gangJoin), soaking up a
// trampler's overflow as they do.
//
// It is an estimate of a player trying to lose as little as possible,
// and to kill what it can for free, not a proof: where the greedy
// second step cannot get the defender
// below lethal but unblockedPower — exact without trample, a lower
// bound with it — says it might live, the answer falls back to the
// pessimistic one: the defender loses nothing and every attacker is
// blocked (guessed).
func (p *Policy) blockToSurvive(st *state, def *SeatEval, swing, blockers []*protocol.CardView) defence {
	d := defence{
		deadDef:     map[string]bool{},
		blockedMine: map[string]bool{},
		blockedBy:   map[string][]*protocol.CardView{},
	}
	survives := func(a, b *protocol.CardView) bool {
		return couldBlock(st, def.ID, a, b) && !blockerDies(a, b)
	}
	_, blockerOf := matchBlocks(swing, blockers, survives)
	used := make([]bool, len(blockers))
	block := func(a *protocol.CardView, bi int) {
		used[bi] = true
		d.blockedMine[a.InstanceID] = true
		d.blockedBy[a.InstanceID] = append(d.blockedBy[a.InstanceID], blockers[bi])
	}
	done := func() defence {
		for bi, b := range blockers {
			if !used[bi] {
				d.spare = append(d.spare, b)
			}
		}
		return d
	}
	// left is what each attacker still connects for under the blocks
	// chosen so far. A blocked attacker connects for nothing — unless
	// it tramples (#1504), when it connects for whatever its blockers
	// do not absorb. A block the blocker survives usually absorbs all
	// of a trampler, but not when it survives by being indestructible
	// or protected: those take lethal assignment like anything else,
	// and the rest goes over.
	left := make([]int, len(swing))
	for ai, a := range swing {
		left[ai] = a.Power
		if bi := blockerOf[ai]; bi >= 0 {
			block(a, bi)
			left[ai] = overflowPast(a, blockers[bi], left[ai])
		}
		d.through += left[ai]
	}
	order := byPower(swing)
	// The defender gang-blocks to kill (#1548). The minimal-losses
	// blocks stop the moment the defender would live, but a block that
	// does not kill its attacker can often be made one by adding spare
	// blockers to it for next to nothing — a second Ogre on a chumped
	// Wurm costs nothing, since the Wurm's 7 kills one Ogre either way.
	// decideBlock makes that block, so the model does too: biggest
	// attacker first, the cheapest join that kills it, whenever the
	// kill is worth what the join costs (gangJoin). The joiners also
	// soak up a trampler's overflow, so NOW falls by what they absorb.
	//
	// The defender's dead are recounted the way this bot's damage is
	// assigned: down the blockers in the order they were declared,
	// lethal to each while it lasts (orderedKills) — the chump first,
	// then the joiners. A second Ogre on a chumped Wurm leaves the
	// first Ogre dead and the second alive.
	gangUp := func() {
		for _, ai := range order {
			a := swing[ai]
			if !d.blockedMine[a.InstanceID] {
				continue
			}
			var cands []*protocol.CardView
			var at []int
			for bi, b := range blockers {
				if !used[bi] && couldBlock(st, def.ID, a, b) {
					cands = append(cands, b)
					at = append(at, bi)
				}
			}
			join, ok := st.gangJoin(a, d.blockedBy[a.InstanceID], cands)
			if !ok {
				continue
			}
			for _, j := range join {
				block(a, at[j])
			}
			gang := d.blockedBy[a.InstanceID]
			for _, b := range gang {
				delete(d.deadDef, b.InstanceID)
			}
			for _, b := range orderedKills(a, gang) {
				d.deadDef[b.InstanceID] = true
			}
			over := connects(a, gang)
			d.through -= left[ai] - over
			left[ai] = over
		}
	}
	if d.through < def.Life {
		gangUp()
		return done()
	}
	// Chump or trade until the defender would live: each time the one
	// block that saves the most, biggest attacker first on a tie, with
	// the least the defender can spare. An attacker without trample
	// takes one blocker and stops; a trampler can take a second and a
	// third while its overflow is still what kills.
	for d.through >= def.Life {
		bestA, bestB, bestSave := -1, -1, 0
		for _, ai := range order {
			if left[ai] <= 0 {
				continue
			}
			a := swing[ai]
			for bi, b := range blockers {
				if used[bi] || !couldBlock(st, def.ID, a, b) {
					continue
				}
				save := left[ai] - overflowPast(a, b, left[ai])
				switch {
				case save > bestSave:
				case save == bestSave && ai == bestA && cheaperBlocker(st, b, blockers[bestB]):
				default:
					continue
				}
				bestA, bestB, bestSave = ai, bi, save
			}
		}
		if bestA < 0 {
			break
		}
		a, b := swing[bestA], blockers[bestB]
		block(a, bestB)
		// The blocker dies only when it is assigned lethal damage: all
		// of a non-trampler's power, or its share of a trampler's —
		// and not when it strikes first at a trampler, since that
		// block is counted as killing the trampler before it deals any.
		dies := blockerDies(a, b)
		if hasKeyword(a, "trample") {
			need := effectiveToughness(b)
			if hasKeyword(a, "deathtouch") {
				need = 1
			}
			dies = dies && bestSave >= need && !(firstStrikes(b) && !firstStrikes(a))
		}
		if dies {
			d.deadDef[b.InstanceID] = true
		}
		left[bestA] -= bestSave
		d.through -= bestSave
	}
	if d.through >= def.Life {
		best := unblockedPower(st, def.ID, swing, blockers)
		if best < def.Life {
			// The greedy missed a way to live that the defender may
			// have: assume it finds one and loses nothing doing so.
			g := defence{through: best, deadDef: map[string]bool{}, blockedMine: map[string]bool{}, guessed: true}
			for _, a := range swing {
				g.blockedMine[a.InstanceID] = true
			}
			return g
		}
	}
	gangUp()
	return done()
}

// cheaperBlocker is the defender's tie-break between two blocks that
// save the same damage: the smaller body, then the one it values less.
func cheaperBlocker(st *state, b, than *protocol.CardView) bool {
	if b.Power != than.Power {
		return b.Power < than.Power
	}
	return st.w.CombatValue(b) < st.w.CombatValue(than)
}

// crackBackPower is unblockedPower from the other side of the table,
// made pessimistic where unblockedPower is optimistic: it is the
// answer to "can this kill me?", so every simplification has to err
// toward yes. A trampler is held only by a blocker that absorbs all of
// its power (and a deathtouch trampler by nothing — CR 702.19c lets it
// assign one point per blocker), and menace counts as unblockable
// rather than as needing two blockers.
func crackBackPower(st *state, attackers, blockers []*protocol.CardView) int {
	through, _ := matchBlocks(attackers, blockers, func(a, b *protocol.CardView) bool {
		if !couldBlock(st, st.me, a, b) || hasKeyword(a, "menace") {
			return false
		}
		if hasKeyword(a, "trample") {
			return !hasKeyword(a, "deathtouch") && effectiveToughness(b) >= a.Power
		}
		return true
	})
	return through
}

// raceAttack declares the next attacker of a committed race, or reports
// that every attacker in it has been declared — in which case the bot
// is done attacking: the creatures outside the swing are the reserve
// the crack-back check counted on.
func (p *Policy) raceAttack(moves []legal.Move, plan *racePlan) (aiseat.Decision, bool) {
	want := make(map[string]int, len(plan.swing))
	for i, c := range plan.swing {
		want[c.InstanceID] = i
	}
	best, bestRank := -1, 0
	for i := range moves {
		if moves[i].Kind != legal.KindAttack {
			continue
		}
		ap := decode[attackParams](moves[i].Params)
		rank, ok := want[ap.Attacker]
		if !ok || ap.Target != plan.target {
			continue
		}
		if best < 0 || rank < bestRank {
			best, bestRank = i, rank
		}
	}
	if best < 0 {
		return aiseat.Decision{}, false
	}
	reason := plan.reason
	if reason == "" {
		reason = fmt.Sprintf("attack: two-turn race — %d now + %d next turn ≥ their %d life; crack-back %d < my %d",
			plan.now, plan.next, plan.life, plan.crack, plan.myLife)
	}
	return aiseat.Decision{Index: best, Reason: reason}, true
}
