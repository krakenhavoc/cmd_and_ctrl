package heuristic

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// attrition.go is the race's longer horizon (#1548, the second option
// #1527 named).
//
// The two-turn race (race.go) cashes an edge that shows within two
// turns. Once the defender gang-blocks to kill, plenty of edges stop
// showing that fast. The full seed-1409 Wurm-edge mirror is the board:
// two Wurms, six Ogres, five Bears and a 3/3 commander against one Wurm
// and the rest the same, both players on 6. A Wurm into the Ogres is
// chumped and the next Ogre joins to kill it, and whatever I send, the
// best two-turn line puts 2 through against 6 life. But one Wurm into
// their Wurm is an even trade that keeps the edge, and the board after
// it races: my last Wurm and six Ogres into six Ogres, five Bears and a
// commander is nothing now and 6 next turn.
//
// So when no race exists the bot looks further, by playing the race's
// own model forward. An exchange is acceptable when
//
//   - the crack-back on the turn in between is less than my life: the
//     race's own safety check, with each turn's crack-back taken off my
//     life for the next;
//   - it makes progress the defender cannot avoid for free: damage it
//     can stop only by losing a creature, or a creature of theirs that
//     dies and stays dead (a commander goes back to the command zone
//     and is cast again, so neither its blocks nor its death count —
//     freeBlock in race.go);
//   - it is even or favourable: no more of my creatures die than of
//     theirs;
//   - and the edge survives it: what I have left is still worth more
//     than what they have left (CombatValue).
//
// After an acceptable exchange the board is rebuilt with the dead
// removed and both sides' creatures untapped and undamaged — mine
// untap on my turn, and the defender's are assumed home to block,
// which is the pessimistic way round — and asked again: a lethal all-in
// or a two-turn race ends the line, and otherwise one more acceptable
// exchange, up to attritionHorizon of them. The swing the bot sends is
// the first exchange of a line that ends in a race, tried in the race's
// own order and chosen the race's way (#2400): with
// Config.RaceEndsSoonest the line with the fewest exchanges, then the
// most sure damage, then the smallest swing; without it the first line
// found, which is the smallest swing that works.
//
// It inherits the race's pessimism and its model of the defender:
// block to survive losing as little as possible, gang-block to kill.
// Nothing is remembered between decisions. A defender that does
// something else only changes what the next turn's search finds.
//
// The search is bounded by attritionBudget swings evaluated per pass,
// across every defender and depth, so a wide board costs the same few
// milliseconds as a narrow one; running out finds nothing, which leaves
// the bot exactly where it was without the horizon. There are at most
// two passes: the second, with Config.HorizonRecastsCommander, runs
// only when the first finds nothing (planAttrition).
const (
	// attritionHorizon is how many acceptable exchanges a line may make
	// before the race that ends it.
	attritionHorizon = 3
	// attritionBudget is how many swings one decision's search may
	// evaluate. The full Wurm mirror's line costs 8; a swing is tens of
	// microseconds on a board of a dozen creatures a side.
	attritionBudget = 100
)

// attritionStep is the first exchange of a line that ends in a race.
type attritionStep struct {
	swing []*protocol.CardView
	e     raceEval
	x     exchange
	// turns counts the exchanges in the line, this one included,
	// before the race.
	turns int
}

// search is one pass of the attrition search: the swings it may still
// evaluate, and whether a commander of mine that dies in an exchange is
// cast again (recast, Config.HorizonRecastsCommander's pass).
type search struct {
	budget int
	recast bool
}

// planAttrition looks for a line of acceptable exchanges ending in a
// race against any opponent, focus first. Nil when there is none.
//
// With Config.HorizonRecastsCommander (#2474) a second pass, with a
// budget of its own, looks again with a commander of mine that dies
// cast again — only when the first finds nothing. Allowed in the first
// pass, the trade made more exchanges acceptable at every depth, and
// the search spent its budget on them before it reached lines it used
// to find (gate seeds 101, 103 and 111). As a second pass it changes
// no decision the search made before, and only adds lines where there
// were none.
func (p *Policy) planAttrition(st *state, moves []legal.Move, focus string) *racePlan {
	if st.myEval == nil || st.myEval.Life <= 0 {
		return nil
	}
	if plan := p.planAttritionWith(st, moves, focus, &search{budget: attritionBudget}); plan != nil || !p.cfg.HorizonRecastsCommander {
		return plan
	}
	return p.planAttritionWith(st, moves, focus, &search{budget: attritionBudget, recast: true})
}

// planAttritionWith is one pass of planAttrition.
func (p *Policy) planAttritionWith(st *state, moves []legal.Move, focus string, s *search) *racePlan {
	for _, def := range focusFirst(st.opps, focus) {
		if def.Life <= 0 || def.CantLoseLife {
			continue
		}
		committed, joinable := raceSwing(st, moves, def)
		if len(joinable) == 0 {
			continue
		}
		step := p.attritionLine(st, def, committed, joinable, attritionHorizon, s, p.cfg.RaceEndsSoonest)
		if step == nil {
			continue
		}
		back := ""
		if step.x.back > 0 {
			back = fmt.Sprintf(" (%d cast again)", step.x.back)
		}
		return &racePlan{
			target: def.ID, swing: step.swing,
			now: step.e.now, next: step.e.next, crack: step.e.crack,
			life: def.Life, myLife: st.myEval.Life,
			reason: fmt.Sprintf("attack: attrition — %d of mine%s for %d of theirs and %d sure now, a race within %d exchange(s); crack-back %d < my %d",
				step.x.mine, back, step.x.theirs, step.x.sure, step.turns, step.x.crack, st.myEval.Life),
		}
	}
	return nil
}

// attritionLine tries the prefixes of the attack order, as raceWith
// does, and returns the first whose exchange is acceptable and leaves a
// board that races within depth more exchanges.
//
// With soonest (Config.RaceEndsSoonest at the top of the search,
// #2400) it tries every prefix and returns the line that ends the game
// soonest (swingKey): the fewest exchanges before the race, then the
// most sure damage, then the smallest swing. The deeper levels only ask
// whether some line exists, so they keep the first. Running out of
// budget returns the best found so far.
func (p *Policy) attritionLine(st *state, def *SeatEval, committed, joinable []*protocol.CardView, depth int, s *search, soonest bool) *attritionStep {
	if depth <= 0 || !hasEdge(st, def) {
		return nil
	}
	blockers := defenderBlockers(st, def.ID)
	first := 1
	if len(committed) > 0 {
		first = 0
	}
	var best *attritionStep
	for k := first; k <= len(joinable); k++ {
		if s.budget <= 0 {
			return best
		}
		s.budget--
		swing := make([]*protocol.CardView, 0, len(committed)+k)
		swing = append(swing, committed...)
		swing = append(swing, joinable[:k]...)
		e := p.raceNumbers(st, def, swing, blockers)
		o := p.outcomeOf(st, def, e, s.recast)
		x := st.exchangeOf(def, e, o)
		if !x.acceptable(st) {
			continue
		}
		next, nextDef := st.afterExchange(def, o, x)
		if !hasEdge(next, nextDef) {
			continue
		}
		turns, ok := p.racesFrom(next, nextDef, depth-1, s)
		if !ok {
			continue
		}
		step := &attritionStep{swing: swing, e: e, x: x, turns: turns + 1}
		if !soonest {
			return step
		}
		if best == nil || step.key().beats(best.key()) {
			best = step
		}
	}
	return best
}

// key ranks the step's line for Config.RaceEndsSoonest (swingKey).
func (s *attritionStep) key() swingKey {
	return swingKey{turns: s.turns, sure: s.x.sure}
}

// exchange is what one swing of an attrition line is counted as. It is
// read off the race's estimate, pessimistically, because the line it
// belongs to is several turns long and a wrong guess repeated every turn
// is a game that never ends.
type exchange struct {
	// sure is the damage the defender cannot stop for free: the
	// least of what it takes blocking to survive and what gets past
	// its free blocks (raceEval.free, the race's free answer). A
	// defender that can block with a creature that lives, or with its
	// commander, takes none of that and loses nothing, and the same
	// swing comes round next turn. Damage it can only stop by losing
	// a creature is progress either way.
	sure int
	// mine and theirs are the creatures that die on each side. A
	// commander of theirs is not counted: it goes back to the command
	// zone and is cast again (CR 903.9a), so trading into it changes
	// nothing. Mine is counted unless I cast it again
	// (outcome.mineBack, #2474); back is how many of mine are.
	mine, theirs int
	back         int
	crack        int
}

// outcome is the answer to one swing that an attrition line plays
// forward: the attackers of mine that die, the ones of those I cast
// again, and the defender's creatures that die.
type outcome struct {
	mineDead, mineBack, theirDead map[string]bool
}

// outcomeOf is the answer the line plays forward. It is
// blockToSurvive's, the race's model of a defender blocking to
// survive.
//
// Where that model only guessed (defence.guessed), it counts every
// attacker of mine blocked and dead and nothing of theirs, so the
// exchange is never acceptable. With Config.HorizonPlaysFreeAnswer
// (#2474) the line plays the free answer instead, when the defender
// survives it: a named way to live that loses it nothing for good, the
// kind of answer the guess stood in for. In the full seed-1409 Wurm
// mirror's last board the greedy blocks my Bear with their commander,
// which lives, and then has nothing left to stop my commander; the
// free answer blocks commander with commander.
//
// With recast (the second pass of Config.HorizonRecastsCommander,
// #2474) a commander of mine that dies is cast again from the command
// zone (CR 903.9a) when my mana sources pay its mana value plus its
// tax (CR 903.8, canRecast), so it is not lost. The defender's
// commander has always been counted that way here.
func (p *Policy) outcomeOf(st *state, def *SeatEval, e raceEval, recast bool) outcome {
	o := outcome{mineDead: e.mineDead, mineBack: map[string]bool{}, theirDead: e.d.deadDef}
	if p.cfg.HorizonPlaysFreeAnswer && e.d.guessed && e.free < def.Life {
		o.mineDead, o.theirDead = e.freeAns.mineDead, map[string]bool{}
	}
	if recast {
		for i := range st.view.Battlefield.Cards {
			c := &st.view.Battlefield.Cards[i]
			if o.mineDead[c.InstanceID] && c.IsCommander && c.Owner == st.me && st.canRecast(st.me, c) {
				o.mineBack[c.InstanceID] = true
			}
		}
	}
	return o
}

// exchangeOf counts e, answered with o, as an exchange.
func (st *state) exchangeOf(def *SeatEval, e raceEval, o outcome) exchange {
	x := exchange{sure: e.sure(), mine: len(o.mineDead) - len(o.mineBack), back: len(o.mineBack), crack: e.crack}
	for i := range st.view.Battlefield.Cards {
		c := &st.view.Battlefield.Cards[i]
		if c.Controller == def.ID && o.theirDead[c.InstanceID] && !c.IsCommander {
			x.theirs++
		}
	}
	return x
}

// acceptable is the per-exchange test: a crack-back I survive, progress
// the defender cannot avoid for free (sure damage, or a creature of
// theirs that stays dead), and no more of my creatures dead than of
// theirs.
func (x exchange) acceptable(st *state) bool {
	if x.crack >= st.myEval.Life {
		return false
	}
	if x.sure <= 0 && x.theirs == 0 {
		return false
	}
	return x.mine <= x.theirs
}

// racesFrom reports whether board st, at the start of my next turn, has
// a lethal all-in or a two-turn race against def — or, with depth left,
// an acceptable exchange that leads to one. turns counts the exchanges
// before it.
//
// In the pass that casts my commander again (#2474), a commander of
// mine cast again after the last exchange is summoning sick on this
// board: it can block, and it counts for a race's next turn, but it
// cannot attack yet. When nothing else works the line may wait a turn
// for it (afterWait), and that turn counts as an exchange.
func (p *Policy) racesFrom(st *state, def *SeatEval, depth int, s *search) (int, bool) {
	if def.Life <= 0 {
		return 0, true
	}
	if st.myEval.Life <= 0 {
		return 0, false
	}
	if joinable := simAttackers(st, def); len(joinable) > 0 {
		if unblockedPower(st, def.ID, joinable, defenderBlockers(st, def.ID)) >= def.Life {
			return 0, true
		}
		if p.raceWith(st, def, nil, joinable, &s.budget) != nil {
			return 0, true
		}
		if step := p.attritionLine(st, def, nil, joinable, depth, s, false); step != nil {
			return step.turns, true
		}
	}
	if !s.recast || depth <= 0 || !st.waitingOn() || s.budget <= 0 {
		return 0, false
	}
	s.budget--
	after, ok := st.afterWait()
	if !ok {
		return 0, false
	}
	if turns, ok := p.racesFrom(after, def, depth-1, s); ok {
		return turns + 1, true
	}
	return 0, false
}

// waitingOn reports whether a creature of mine on a simulated board
// cannot attack yet: a commander cast again after an exchange.
func (st *state) waitingOn() bool {
	for i := range st.view.Battlefield.Cards {
		c := &st.view.Battlefield.Cards[i]
		if c.Controller == st.me && isCreature(c) && sickToAttack(c) {
			return true
		}
	}
	return false
}

// sickToAttack reports whether c cannot attack yet: summoning sick
// without haste (CR 302.6, 702.10).
func sickToAttack(c *protocol.CardView) bool {
	return c.SummoningSick && !hasKeyword(c, "haste")
}

// afterWait is the board at the start of my turn after one on which I
// attack with nothing (#2474): every opponent swings everything at me
// into every creature I have, the race's pessimistic crack-back, and
// nothing of mine is summoning sick any more. Not ok when that
// crack-back kills me. The view is a copy; st is not touched.
func (st *state) afterWait() (*state, bool) {
	var home []*protocol.CardView
	for i := range st.view.Battlefield.Cards {
		c := &st.view.Battlefield.Cards[i]
		if c.Controller == st.me && isCreature(c) {
			home = append(home, c)
		}
	}
	crack := 0
	for _, o := range st.opps {
		if o.Life <= 0 {
			continue
		}
		var theirs []*protocol.CardView
		for i := range st.view.Battlefield.Cards {
			c := &st.view.Battlefield.Cards[i]
			if c.Controller == o.ID && isCreature(c) && !hasKeyword(c, "defender") {
				theirs = append(theirs, c)
			}
		}
		crack += crackBackPower(st, theirs, home)
	}
	if crack >= st.myEval.Life {
		return nil, false
	}
	cards := append([]protocol.CardView(nil), st.view.Battlefield.Cards...)
	for i := range cards {
		if cards[i].Controller == st.me {
			cards[i].SummoningSick = false
		}
	}
	view := *st.view
	view.Battlefield.Cards = cards
	next := *st
	next.view = &view
	me := *st.myEval
	me.Life -= crack
	next.myEval = &me
	return &next, true
}

// hasEdge reports whether my creatures are worth more than def's.
func hasEdge(st *state, def *SeatEval) bool {
	mine, theirs := 0.0, 0.0
	for i := range st.view.Battlefield.Cards {
		c := &st.view.Battlefield.Cards[i]
		if !isCreature(c) {
			continue
		}
		switch c.Controller {
		case st.me:
			mine += st.w.CombatValue(c)
		case def.ID:
			theirs += st.w.CombatValue(c)
		}
	}
	return mine > theirs
}

// afterExchange is the board at the start of my next turn, as the
// attrition search assumes it: every creature o kills gone, except a
// commander of theirs, which is cast again, and a commander of mine o
// casts again (mineBack), which is back summoning sick unless it has
// haste; mine and the defender's untapped, undamaged and out of
// combat; the defender's life down by the damage sure to connect and
// mine by the crack-back. A commander of mine cast again counts one
// more cast from the command zone, so a second trade in the same line
// pays the higher tax. Other seats are left as they are. The view is a
// copy; st is not touched.
func (st *state) afterExchange(def *SeatEval, o outcome, x exchange) (*state, *SeatEval) {
	cards := make([]protocol.CardView, 0, len(st.view.Battlefield.Cards))
	for _, c := range st.view.Battlefield.Cards {
		back := o.mineBack[c.InstanceID]
		if (o.mineDead[c.InstanceID] && !back) || (o.theirDead[c.InstanceID] && !c.IsCommander) {
			continue
		}
		if isCreature(&c) && (c.Controller == st.me || c.Controller == def.ID) {
			c.Tapped, c.SummoningSick = false, back && !hasKeyword(&c, "haste")
			c.AttackingTarget, c.BlockingTarget, c.BlockingTargets = "", "", nil
			c.DamageMarked = 0
		}
		cards = append(cards, c)
	}
	view := *st.view
	view.Battlefield.Cards = cards
	view.Battlefield.Count = len(cards)
	if len(o.mineBack) > 0 {
		view.Seats = append([]protocol.PlayerView(nil), st.view.Seats...)
		for i := range view.Seats {
			if view.Seats[i].ID != st.me {
				continue
			}
			casts := make(map[string]int, len(view.Seats[i].CommanderCasts)+len(o.mineBack))
			for id, n := range view.Seats[i].CommanderCasts {
				casts[id] = n
			}
			for id := range o.mineBack {
				casts[id]++
			}
			view.Seats[i].CommanderCasts = casts
		}
	}
	next := *st
	next.view = &view
	nd := *def
	nd.Life -= x.sure
	next.opps = make([]*SeatEval, len(st.opps))
	for i, o := range st.opps {
		next.opps[i] = o
		if o.ID == def.ID {
			next.opps[i] = &nd
		}
	}
	me := *st.myEval
	me.Life -= x.crack
	next.myEval = &me
	return &next, &nd
}

// simAttackers is every creature of mine on a simulated board that
// could attack def, in the race's order. There is no move list to read
// it off, so it keeps out what the engine would: a defender, a
// creature carrying a cant_attack restriction, and one that is
// summoning sick (a commander of mine cast again, #2474).
func simAttackers(st *state, def *SeatEval) []*protocol.CardView {
	var out []*protocol.CardView
	for i := range st.view.Battlefield.Cards {
		c := &st.view.Battlefield.Cards[i]
		if c.Controller != st.me || !isCreature(c) || hasKeyword(c, "defender") || hasRestriction(c, "cant_attack") || sickToAttack(c) {
			continue
		}
		out = append(out, c)
	}
	evasiveFirst(st, def, out)
	return out
}
