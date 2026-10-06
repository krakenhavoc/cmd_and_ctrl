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
// the first exchange of the first line that ends in a race, tried in
// the race's own order, so it too is the smallest that works.
//
// It inherits the race's pessimism and its model of the defender:
// block to survive losing as little as possible, gang-block to kill.
// Nothing is remembered between decisions. A defender that does
// something else only changes what the next turn's search finds.
//
// The search is bounded by attritionBudget swings evaluated per
// decision, across every defender and depth, so a wide board costs the
// same few milliseconds as a narrow one; running out finds nothing,
// which leaves the bot exactly where it was without the horizon.
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

// planAttrition looks for a line of acceptable exchanges ending in a
// race against any opponent, focus first. Nil when there is none.
func (p *Policy) planAttrition(st *state, moves []legal.Move, focus string) *racePlan {
	if st.myEval == nil || st.myEval.Life <= 0 {
		return nil
	}
	budget := attritionBudget
	for _, def := range focusFirst(st.opps, focus) {
		if def.Life <= 0 || def.CantLoseLife {
			continue
		}
		committed, joinable := raceSwing(st, moves, def)
		if len(joinable) == 0 {
			continue
		}
		step := p.attritionLine(st, def, committed, joinable, attritionHorizon, &budget)
		if step == nil {
			continue
		}
		return &racePlan{
			target: def.ID, swing: step.swing,
			now: step.e.now, next: step.e.next, crack: step.e.crack,
			life: def.Life, myLife: st.myEval.Life,
			reason: fmt.Sprintf("attack: attrition — %d of mine for %d of theirs and %d sure now, a race within %d exchange(s); crack-back %d < my %d",
				step.x.mine, step.x.theirs, step.x.sure, step.turns, step.x.crack, st.myEval.Life),
		}
	}
	return nil
}

// attritionLine tries the prefixes of the attack order, as raceWith
// does, and returns the first whose exchange is acceptable and leaves a
// board that races within depth more exchanges.
func (p *Policy) attritionLine(st *state, def *SeatEval, committed, joinable []*protocol.CardView, depth int, budget *int) *attritionStep {
	if depth <= 0 || !hasEdge(st, def) {
		return nil
	}
	blockers := defenderBlockers(st, def.ID)
	first := 1
	if len(committed) > 0 {
		first = 0
	}
	for k := first; k <= len(joinable); k++ {
		if *budget <= 0 {
			return nil
		}
		*budget--
		swing := make([]*protocol.CardView, 0, len(committed)+k)
		swing = append(swing, committed...)
		swing = append(swing, joinable[:k]...)
		e := p.raceNumbers(st, def, swing, blockers)
		x := st.exchangeOf(def, e)
		if !x.acceptable(st) {
			continue
		}
		next, nextDef := st.afterExchange(def, e, x)
		if !hasEdge(next, nextDef) {
			continue
		}
		if turns, ok := p.racesFrom(next, nextDef, depth-1, budget); ok {
			return &attritionStep{swing: swing, e: e, x: x, turns: turns + 1}
		}
	}
	return nil
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
	// nothing. Mine is counted, because the search may not rely on
	// recasting it.
	mine, theirs int
	crack        int
}

// exchangeOf counts e as an exchange.
func (st *state) exchangeOf(def *SeatEval, e raceEval) exchange {
	x := exchange{sure: min(e.d.through, e.free), mine: len(e.mineDead), crack: e.crack}
	for i := range st.view.Battlefield.Cards {
		c := &st.view.Battlefield.Cards[i]
		if c.Controller == def.ID && e.d.deadDef[c.InstanceID] && !c.IsCommander {
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
func (p *Policy) racesFrom(st *state, def *SeatEval, depth int, budget *int) (int, bool) {
	if def.Life <= 0 {
		return 0, true
	}
	if st.myEval.Life <= 0 {
		return 0, false
	}
	joinable := simAttackers(st, def)
	if len(joinable) == 0 {
		return 0, false
	}
	if unblockedPower(st, def.ID, joinable, defenderBlockers(st, def.ID)) >= def.Life {
		return 0, true
	}
	if p.raceWith(st, def, nil, joinable, budget) != nil {
		return 0, true
	}
	if step := p.attritionLine(st, def, nil, joinable, depth, budget); step != nil {
		return step.turns, true
	}
	return 0, false
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
// attrition search assumes it: every creature e kills gone except a
// commander of theirs, which is cast again; mine and the defender's
// untapped, undamaged and out of combat; the defender's life down by
// the damage sure to connect and mine by the crack-back. Other seats
// are left as they are. The view is a copy; st is not touched.
func (st *state) afterExchange(def *SeatEval, e raceEval, x exchange) (*state, *SeatEval) {
	cards := make([]protocol.CardView, 0, len(st.view.Battlefield.Cards))
	for _, c := range st.view.Battlefield.Cards {
		if e.mineDead[c.InstanceID] || (e.d.deadDef[c.InstanceID] && !c.IsCommander) {
			continue
		}
		if isCreature(&c) && (c.Controller == st.me || c.Controller == def.ID) {
			c.Tapped, c.SummoningSick = false, false
			c.AttackingTarget, c.BlockingTarget, c.BlockingTargets = "", "", nil
			c.DamageMarked = 0
		}
		cards = append(cards, c)
	}
	view := *st.view
	view.Battlefield.Cards = cards
	view.Battlefield.Count = len(cards)
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
// it off, so it keeps out what the engine would: a defender, and a
// creature carrying a cant_attack restriction.
func simAttackers(st *state, def *SeatEval) []*protocol.CardView {
	var out []*protocol.CardView
	for i := range st.view.Battlefield.Cards {
		c := &st.view.Battlefield.Cards[i]
		if c.Controller != st.me || !isCreature(c) || hasKeyword(c, "defender") || hasRestriction(c, "cant_attack") {
			continue
		}
		out = append(out, c)
	}
	evasiveFirst(st, def, out)
	return out
}
