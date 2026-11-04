package game

import "github.com/google/uuid"

// faced_dilemma.go — #4973 (Star Trek): "you face a dilemma as you
// choose one or more modes for a spell or ability" (Jett Reno, Problem
// Solver). The Comprehensive Rules have no entry for it; the reminder
// text defines it as the act of choosing modes (CR 700.2a for a spell
// or activated ability, CR 700.2b for a triggered one). The engine
// already has one place for each of those choices, and each emits
// EventFacedDilemma once the choice is made.
//
// Also the extra-turn count H.M.S. Bounty reads, which shipped in the
// same batch: Player.ExtraTurnsBegun.

// emitFacedDilemmaLocked tells the listeners — the trigger harvester
// among them — that `player` chose one or more modes for the spell or
// ability `source` (the spell, or the ability's source).
//
// The event is delivered to the listeners and NOT appended to
// g.Events. It is a silent kind (the cast, activation or mode-choice
// line already says it, protocol's log gate), and keeping it off the
// log is what keeps every game that chooses a mode recording the same
// events it did before the kind existed: the frozen v9 restore points
// (#522; modal_trigger.json is one) hold those event lists, and a
// fixture is never rewritten. It carries the Seq and batch of the event
// before it, so batch-keyed triggers group it with the action that
// chose the modes.
//
// Caller must hold g.mu in write mode.
func (g *Game) emitFacedDilemmaLocked(player, source uuid.UUID) {
	if player == uuid.Nil {
		return
	}
	ev := Event{
		Kind:   EventFacedDilemma,
		Actor:  player,
		Source: source,
		CardID: source,
		Seq:    g.eventSeq,
		Batch:  g.currentEventBatchLocked(),
	}
	g.notifyListenersLocked(ev)
}

// ExtraTurnsBegunForEffect is how many extra turns `player` has begun
// this game (Player.ExtraTurnsBegun): "if you … haven't begun an extra
// turn this game" is a zero here. A player not in the game reads 0.
//
// Caller must hold g.mu.
func (g *Game) ExtraTurnsBegunForEffect(player uuid.UUID) int {
	if p := g.playerByIDLocked(player); p != nil {
		return p.ExtraTurnsBegun
	}
	return 0
}
