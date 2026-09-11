package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"

// concede.go implements aiseat.Conceder.
//
// This is the one heuristic in the package tuned by feeling rather
// than by arithmetic, and it is tuned CONSERVATIVE on purpose. The
// S31 issue says so outright ("default conservative so humans get the
// finish"), and it is the right call twice over: a bot that scoops
// the moment it falls behind robs a human of the win they were
// playing for, and in Commander a seat at 2 life with a board is
// routinely not dead. The cost of conceding a turn too late is that
// the table waits thirty seconds. The cost of conceding a turn too
// early is that somebody's game got taken away from them.
//
// So every one of these has to hold, on three consecutive turns:
//
//   - at or below ConcedeLife (3 by default)
//   - nothing in hand — no upswing left to draw into a hand full of
//     answers, because there is no hand
//   - no creatures on the battlefield
//   - at least one opponent with creatures, i.e. somebody who can
//     actually finish it
//
// A bot at 2 life holding a fog and a wrath does not concede. A bot
// at 2 life with a 5/5 does not concede. A bot at 2 life with nothing
// at all, three turns running, while an opponent has a board, has
// lost, and pretending otherwise wastes everybody's evening.

// ShouldConcede reports whether the position has been hopeless for
// long enough to scoop. It satisfies aiseat.Conceder; a runner that
// does not know about the interface simply never calls it, and the
// policy plays on.
func (p *Policy) ShouldConcede(in aiseat.Input) bool {
	if !p.cfg.Concede {
		return false
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	st := p.newState(in)
	if !p.hopeless(st) {
		p.hopelessTurns, p.hopelessTurn = 0, 0
		return false
	}
	// One turn counts once, however many decisions it contains.
	if st.turn != p.hopelessTurn {
		p.hopelessTurns++
		p.hopelessTurn = st.turn
	}
	return p.hopelessTurns >= p.cfg.ConcedeTurns
}

// HopelessTurns is the number of consecutive turns the position has
// been judged lost. Exposed for tests and for the reasoning line a
// future "show bot reasoning" setting will surface.
func (p *Policy) HopelessTurns() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.hopelessTurns
}

// hopeless is the single-turn judgement. Caller holds p.mu.
func (p *Policy) hopeless(st *state) bool {
	me := st.myEval
	if me == nil || me.Eliminated {
		return false
	}
	if len(st.opps) == 0 {
		// Last seat standing is not a position to concede from.
		return false
	}
	if me.Life > p.cfg.ConcedeLife {
		return false
	}
	if me.Hand > 0 || me.CreatureCount > 0 {
		return false
	}
	// Somebody has to be able to finish it.
	for _, o := range st.opps {
		if o.CreatureCount > 0 {
			return true
		}
	}
	return false
}
