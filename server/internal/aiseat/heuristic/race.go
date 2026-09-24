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
// lives into NEXT when neither the blockers the defender's model put in
// front of it nor any one blocker it left spare can kill it (survives)
// — the defender is assumed to block to kill whenever it can. It is
// still never home: the crack-back check does not change.
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

// racePlan is a committed two-turn race against one defender.
type racePlan struct {
	target string
	// swing is every attacker the plan sends at target, in the order
	// it should declare them — already-declared ones included.
	swing []*protocol.CardView
	// now, next and crack are the three numbers the plan was accepted
	// on, kept for the decision's Reason.
	now, next, crack int
	life, myLife     int
}

// planRace looks for a two-turn race against any opponent, focus first.
// Nil when there is none, which is every board without an edge.
func (p *Policy) planRace(st *state, moves []legal.Move, focus string) *racePlan {
	if st.myEval == nil || st.myEval.Life <= 0 {
		return nil
	}
	order := make([]*SeatEval, 0, len(st.opps))
	for _, o := range st.opps {
		if o.ID == focus {
			order = append([]*SeatEval{o}, order...)
			continue
		}
		order = append(order, o)
	}
	for _, def := range order {
		if def.Life <= 0 {
			continue
		}
		if plan := p.raceAgainst(st, moves, def); plan != nil {
			return plan
		}
	}
	return nil
}

// raceAgainst tries the prefixes of the evasion-sorted attack order
// against one defender and returns the first that races.
func (p *Policy) raceAgainst(st *state, moves []legal.Move, def *SeatEval) *racePlan {
	// Already committed to this defender: part of every swing.
	var committed []*protocol.CardView
	for i := range st.view.Battlefield.Cards {
		c := &st.view.Battlefield.Cards[i]
		if c.Controller == st.me && isCreature(c) && c.AttackingTarget == def.ID {
			committed = append(committed, c)
		}
	}
	// Still able to join, read off the move list so every restriction
	// and every unpayable attack tax the enumerator applied is honoured.
	seen := map[string]bool{}
	var joinable []*protocol.CardView
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
	if len(joinable) == 0 {
		return nil
	}
	blockers := defenderBlockers(st, def.ID)
	answers := func(a *protocol.CardView) int {
		n := 0
		for _, b := range blockers {
			if couldBlock(st, def.ID, a, b) {
				n++
			}
		}
		return n
	}
	// Evasive first, then biggest; stable, so board order breaks ties
	// and the plan never depends on a sort's whim.
	sort.SliceStable(joinable, func(i, j int) bool {
		ai, aj := answers(joinable[i]), answers(joinable[j])
		if ai != aj {
			return ai < aj
		}
		return joinable[i].Power > joinable[j].Power
	})
	// The empty prefix is tried when something is already declared: a
	// swing that already races needs nobody else, and adding to it only
	// spends the reserve.
	first := 1
	if len(committed) > 0 {
		first = 0
	}
	for k := first; k <= len(joinable); k++ {
		swing := make([]*protocol.CardView, 0, len(committed)+k)
		swing = append(swing, committed...)
		swing = append(swing, joinable[:k]...)
		now, next, crack := p.raceNumbers(st, def, swing, blockers)
		if now+next >= def.Life && crack < st.myEval.Life {
			return &racePlan{
				target: def.ID, swing: swing,
				now: now, next: next, crack: crack,
				life: def.Life, myLife: st.myEval.Life,
			}
		}
	}
	return nil
}

// raceNumbers is the three-number estimate for one swing at def.
func (p *Policy) raceNumbers(st *state, def *SeatEval, swing, blockers []*protocol.CardView) (now, next, crack int) {
	d := p.blockToSurvive(st, def, swing, blockers)
	now = d.through

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
		if d.blockedMine[c.InstanceID] {
			// Counted dead unless no block the defender could make
			// kills it (#1527): a Wurm chumped by the last Bear
			// tramples over again next turn.
			if d.survives(st, def.ID, c) {
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
	// defender this swing already killed.
	for _, o := range st.opps {
		if o.ID == def.ID && now >= def.Life {
			continue
		}
		var theirs []*protocol.CardView
		for i := range st.view.Battlefield.Cards {
			c := &st.view.Battlefield.Cards[i]
			if c.Controller != o.ID || !isCreature(c) || hasKeyword(c, "defender") {
				continue
			}
			if o.ID == def.ID && d.deadDef[c.InstanceID] {
				continue
			}
			theirs = append(theirs, c)
		}
		crack += crackBackPower(st, theirs, home)
	}

	// NEXT: my survivors into everything the defender kept.
	var kept []*protocol.CardView
	for i := range st.view.Battlefield.Cards {
		c := &st.view.Battlefield.Cards[i]
		if c.Controller == def.ID && isCreature(c) && !d.deadDef[c.InstanceID] {
			kept = append(kept, c)
		}
	}
	if now < def.Life {
		next = unblockedPower(st, def.ID, second, kept)
	}
	return now, next, crack
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
// It says yes only when neither of these kills it:
//
//   - the blockers the model put in front of it (blockersKill), and
//   - any one blocker the model left spare that could block it instead.
//     The model picks the cheapest chump, and a defender that has a
//     creature able to kill the attacker in that spot is assumed to
//     use it: it blocks to kill when it can.
//
// Where the model only guessed at the blocks, nothing survives.
//
// What it does not ask is whether the defender could add spare
// blockers to a block it is already making, ganging up until their
// damage adds up to the attacker's toughness. The rest of the model
// blocks one-for-one too (matchBlocks, blockToSurvive's greedy), and
// so does the heuristic's own decideBlock, which credits a gang block
// with a kill only when the blocker it adds kills alone. Counting gang
// blocks here would put every blocked attacker in front of a wide
// board back in the grave — the full seed-1409 mirror's spare Wurm,
// chumped by one Ogre with five more behind it, is exactly that board.
func (d *defence) survives(st *state, defender string, a *protocol.CardView) bool {
	if d.guessed {
		return false
	}
	if blockersKill(d.blockedBy[a.InstanceID], a) {
		return false
	}
	for _, b := range d.spare {
		if couldBlock(st, defender, a, b) && blockerKills(b, a) {
			return false
		}
	}
	return true
}

// blockersKill reports whether these blockers, together, kill attacker
// a. One blocker is asked exactly (blockerKills). Several are asked
// pessimistically: any deathtouch among them kills, and so does their
// combined damage reaching a's toughness, a double striker's power
// counted twice — ignoring what a's own first strike would kill before
// they deal damage and what a's protection would prevent, both of
// which only ever make it say "dead" about an attacker that lives.
func blockersKill(blockers []*protocol.CardView, a *protocol.CardView) bool {
	if len(blockers) == 1 {
		return blockerKills(blockers[0], a)
	}
	if hasKeyword(a, "indestructible") {
		return false
	}
	total := 0
	for _, b := range blockers {
		if b.Power <= 0 {
			continue
		}
		if hasKeyword(b, "deathtouch") {
			return true
		}
		total += combatDamage(b)
	}
	return total >= effectiveToughness(a)
}

// blockerKills reports whether blocker b, blocking a alone, kills it.
// It is kills with the two timing rules kills does not read:
//
//   - a double striker deals its power twice (CR 702.4b), so a 2/2
//     double striker kills a 4/4;
//   - an attacker with first strike that kills a blocker without it
//     does so in the first combat damage step, and the blocker, gone,
//     deals no damage in the second (CR 510.4).
func blockerKills(b, a *protocol.CardView) bool {
	if hasKeyword(a, "indestructible") || b.Power <= 0 || protectedFrom(a, b) {
		return false
	}
	if firstStrikes(a) && !firstStrikes(b) && kills(a, b) {
		return false
	}
	if hasKeyword(b, "deathtouch") {
		return true
	}
	return combatDamage(b) >= effectiveToughness(a)
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
// only the chump's toughness.
//
// It is an estimate of a player trying to lose as little as possible,
// not a proof: where the greedy second step cannot get the defender
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
		return couldBlock(st, def.ID, a, b) && !kills(a, b)
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
	if d.through < def.Life {
		return done()
	}
	// Chump or trade until the defender would live: each time the one
	// block that saves the most, biggest attacker first on a tie, with
	// the least the defender can spare. An attacker without trample
	// takes one blocker and stops; a trampler can take a second and a
	// third while its overflow is still what kills.
	order := byPower(swing)
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
		dies := kills(a, b)
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
	return aiseat.Decision{
		Index: best,
		Reason: fmt.Sprintf("attack: two-turn race — %d now + %d next turn ≥ their %d life; crack-back %d < my %d",
			plan.now, plan.next, plan.life, plan.crack, plan.myLife),
	}, true
}
