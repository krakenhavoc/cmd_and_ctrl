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
// estimate has: every one of my blocked attackers is counted dead, the
// defender keeps every creature it did not have to lose, it is assumed
// to both attack with all of them AND keep all of them home, every
// opponent swings at me rather than at each other, a trampler is only
// held by a blocker that can absorb all of it, and a menace creature
// counts as unblockable. An estimate that is wrong errs toward not
// racing, which leaves the bot exactly where it was before #1409.
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
	now, deadDef, blockedMine := p.blockToSurvive(st, def, swing, blockers)

	inSwing := make(map[string]bool, len(swing))
	for _, c := range swing {
		inSwing[c.InstanceID] = true
	}
	// Home: what I will have untapped on the turn in between. A swing
	// attacker stays untapped only with vigilance, and only if it lived.
	var home, second []*protocol.CardView
	for i := range st.view.Battlefield.Cards {
		c := &st.view.Battlefield.Cards[i]
		if c.Controller != st.me || !isCreature(c) {
			continue
		}
		if blockedMine[c.InstanceID] {
			continue // counted dead: every blocked attacker is
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
			if o.ID == def.ID && deadDef[c.InstanceID] {
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
		if c.Controller == def.ID && isCreature(c) && !deadDef[c.InstanceID] {
			kept = append(kept, c)
		}
	}
	if now < def.Life {
		next = unblockedPower(st, def.ID, second, kept)
	}
	return now, next, crack
}

// blockToSurvive is the defender's answer to a swing: first every
// block it survives (those cost it nothing), and only if the damage
// still getting through would kill it, the cheapest blocks it does not
// survive, biggest attacker first, until it would live. It returns the
// damage that connects, the defender's creatures that die blocking,
// and every attacker of mine that was blocked at all.
//
// It is an estimate of a player trying to lose as little as possible,
// not a proof: where the greedy second step cannot get the defender
// below lethal but the exact matching (unblockedPower) can, the answer
// falls back to the pessimistic one — the defender loses nothing and
// every attacker is blocked.
func (p *Policy) blockToSurvive(st *state, def *SeatEval, swing, blockers []*protocol.CardView) (through int, deadDef, blockedMine map[string]bool) {
	deadDef, blockedMine = map[string]bool{}, map[string]bool{}
	survives := func(a, b *protocol.CardView) bool {
		return couldBlock(st, def.ID, a, b) && !kills(a, b)
	}
	through, blockerOf := matchBlocks(swing, blockers, survives)
	used := make([]bool, len(blockers))
	for ai, bi := range blockerOf {
		if bi >= 0 {
			used[bi] = true
			blockedMine[swing[ai].InstanceID] = true
		}
	}
	if through < def.Life {
		return through, deadDef, blockedMine
	}
	// Chump or trade, biggest unblocked attacker first, with the least
	// the defender can spare.
	order := byPower(swing)
	for _, ai := range order {
		if through < def.Life {
			break
		}
		if blockerOf[ai] >= 0 {
			continue
		}
		a := swing[ai]
		pick := -1
		for bi, b := range blockers {
			if used[bi] || !couldBlock(st, def.ID, a, b) {
				continue
			}
			if pick < 0 || b.Power < blockers[pick].Power ||
				(b.Power == blockers[pick].Power && st.w.CombatValue(b) < st.w.CombatValue(blockers[pick])) {
				pick = bi
			}
		}
		if pick < 0 {
			continue
		}
		used[pick] = true
		blockedMine[a.InstanceID] = true
		if kills(a, blockers[pick]) {
			deadDef[blockers[pick].InstanceID] = true
		}
		through -= a.Power
	}
	if through >= def.Life {
		exact := unblockedPower(st, def.ID, swing, blockers)
		if exact < def.Life {
			// The greedy missed a way to live that the matching found:
			// assume the defender finds it and loses nothing doing so.
			all := map[string]bool{}
			for _, a := range swing {
				all[a.InstanceID] = true
			}
			return exact, map[string]bool{}, all
		}
	}
	return through, deadDef, blockedMine
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
