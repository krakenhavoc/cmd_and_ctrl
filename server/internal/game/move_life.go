package game

import "github.com/google/uuid"

// move_life.go — life instructions after a replaced move (ADR 0178,
// #2950).
//
// The Darkness Crystal: "If a nontoken creature an opponent controls
// would die, instead exile it and you gain 2 life." The modified event
// is two instructions in printed order (CR 608.2c, 614.6): the exile,
// then the gain. The gain is a life event of its own, with its own
// CR 614 window, so it can pause on a CR 616 ordering prompt (two
// different life replacements on the gaining player, CR 616.1g).
//
// The replacement used to gain the life inside its Replace callback.
// That landed the gain before the card had moved, and when the gain
// paused, the move went on and ran its caller's continuation (the
// "then draw a card" behind a destroy) with the life question still
// open. Now the replacement only DECLARES the gain on the move event
// (ChangeLifeAfterReplacedMoveForEffect). The move's finisher lands the
// move first, then runs each declared life change through the ordinary
// life pipeline, and only after the last one has settled runs the
// move's own continuation (runMoveTailLocked).
//
// The rest of the move rides each life change's ordinary tail as data
// (lifeTail.move, #4361 S9): the instructions not yet run, and a copy of
// the move's route on the life event's own zoneRoute, never the parent
// event or a route pointer. Each run copies the route again before
// running it, and an undo snapshot has its own copy of the event's route
// (cloneReplacementResume), so a replayed answer finds the continuation
// intact rather than consumed by the live game. No new field reaches a
// closure from Game (closure_fields.txt, ADR 0041 P3): the route rides an
// edge the event already has. A life change paused with one is still
// counted by the census (replacementEventKeyable), as before.

// lifeInstruction is one declared life change: who changes, by how much
// (signed, negative for a loss), and the source the log credits. All
// three are fixed as the replacement applies, so a source that leaves
// afterwards does not change who gains (ADR 0178 §4).
type lifeInstruction struct {
	Source uuid.UUID
	Player uuid.UUID
	Delta  int
}

// moveLifeSequence is what is left of a replaced move once the card has
// landed: the life instructions not yet run, and a copy of the move's
// route, whose continuation runs after the last of them. A value: it is
// captured by value and never written after it is built.
type moveLifeSequence struct {
	remaining     []lifeInstruction
	route         zoneRoute
	hasRoute      bool
	mustSettleNow bool
}

// moveLifeRest is a moveLifeSequence riding a life event's tail
// (lifeTail.move): everything but the route, which rides the life
// event's zoneRoute. A value, never written after it is built.
type moveLifeRest struct {
	remaining     []lifeInstruction
	hasRoute      bool
	mustSettleNow bool
}

// sequence is the rest of the move again, with the route taken off the
// life event that carried it. The event's route is cleared, so it cannot
// run twice.
func (r *moveLifeRest) sequence(ev *ReplacementEvent) moveLifeSequence {
	seq := moveLifeSequence{remaining: r.remaining, hasRoute: r.hasRoute, mustSettleNow: r.mustSettleNow}
	if r.hasRoute && ev != nil && ev.zoneRoute != nil {
		seq.route = *ev.zoneRoute
		ev.zoneRoute = nil
	}
	return seq
}

// ChangeLifeAfterReplacedMoveForEffect declares, from a replacement's
// Replace callback, that the modified zone-change event also changes
// player's life by delta after the card has moved ("instead exile it
// and you gain 2 life"). The change runs through the ordinary life
// pipeline, so life replacements apply to it and it may ask a CR 616
// question; the move's continuation waits for it.
//
// The instruction survives a later replacement that redirects or
// cancels the move (CR 614.6: an impossible instruction is ignored, the
// others still happen). A move abandoned before it happens, because its
// prompt was taken away, discards it (CR 614.7).
//
// Returns ErrInvalidParam for an event that is not a zone change.
//
// Caller must hold g.mu (Replace callbacks always do).
func (g *Game) ChangeLifeAfterReplacedMoveForEffect(ev *ReplacementEvent, source, player uuid.UUID, delta int) error {
	if ev == nil || ev.Kind != RepEventMove {
		return ErrInvalidParam
	}
	if delta == 0 || player == uuid.Nil {
		return nil
	}
	// A fresh backing array: an undo snapshot taken while this event was
	// paused shares the old one (cloneReplacementResume copies too).
	n := len(ev.lifeAfterMove)
	ev.lifeAfterMove = append(ev.lifeAfterMove[:n:n], lifeInstruction{Source: source, Player: player, Delta: delta})
	return nil
}

// runMoveTailLocked is the terminal outcome of a zone-change event that
// has reached its finisher: landed, refused or cancelled. With no
// declared life instruction it is runRouteTailLocked. With some, the
// route's continuation is taken off the route and run only after the
// last life change has settled, which may be from a CR 616 resume.
//
// A paired simultaneous entry (Goblin Welder) still lands now, with the
// exit, because it is part of the same event.
//
// Caller must hold g.mu.
func (g *Game) runMoveTailLocked(ev *ReplacementEvent) error {
	if ev == nil {
		return nil
	}
	// ADR 0188 Decision 2: a declared extra turn is queued first, before
	// the life instructions, in printed order. It cannot pause.
	g.queueExtraTurnsAfterMoveLocked(ev)
	if len(ev.lifeAfterMove) == 0 {
		return g.runRouteTailLocked(ev.zoneRoute)
	}
	seq := moveLifeSequence{remaining: ev.lifeAfterMove, mustSettleNow: ev.mustSettleNow}
	ev.lifeAfterMove = nil
	if r := ev.zoneRoute; r != nil {
		if r.hasTail() {
			// #4253 slice 9: a keyed batch's or face-down leg's run is a
			// continuation too, and waits for the life change as `then`
			// does.
			seq.route = r.detachTail()
			seq.route.simultaneousEntry = nil
			seq.hasRoute = true
		}
		if err := g.runRouteTailLocked(r); err != nil {
			return err
		}
	}
	return g.continueMoveLifeLocked(seq)
}

// continueMoveLifeLocked runs the next declared life change, carrying
// the rest of the sequence on its life tail, or, when none is left, the
// move's route continuation. A player who has left the game is skipped
// (CR 800.4a).
//
// Caller must hold g.mu.
func (g *Game) continueMoveLifeLocked(seq moveLifeSequence) error {
	for len(seq.remaining) > 0 {
		in := seq.remaining[0]
		seq.remaining = seq.remaining[1:]
		if p := g.playerByIDLocked(in.Player); p == nil || p.Eliminated {
			continue
		}
		ev := &ReplacementEvent{
			Kind:       RepEventLife,
			Source:     in.Source,
			LifePlayer: in.Player,
			LifeDelta:  in.Delta,
			// ADR 0178 §5: a move paid as a cost cannot pause, and
			// neither can the life change that is part of it.
			mustSettleNow: seq.mustSettleNow,
			lifeTail: &lifeTail{move: &moveLifeRest{
				remaining:     seq.remaining,
				hasRoute:      seq.hasRoute,
				mustSettleNow: seq.mustSettleNow,
			}},
		}
		if seq.hasRoute {
			route := seq.route
			if len(route.simultaneousExit) > 0 {
				route.simultaneousExit = append([]Card(nil), route.simultaneousExit...)
			}
			ev.zoneRoute = &route
		}
		paused, err := g.changeLifeThroughReplacementsLocked(ev)
		if !paused && ev.lifeTail != nil {
			// The change ended on an error without reaching its tail.
			// The rest of the sequence is still owed.
			if tailErr := g.runLifeTailLocked(ev, 0); tailErr != nil && err == nil {
				err = tailErr
			}
		}
		return err
	}
	return g.finishMoveLifeSequenceLocked(seq)
}

// finishMoveLifeSequenceLocked runs the move's route continuation once
// every life instruction has settled, inside the simultaneous exit the
// route carries, as finishBattlefieldLeaveLocked does for an ordinary
// exit. It runs a fresh copy of the route each time, because the tail
// is cleared through the pointer as it runs and an undo replay of the
// same answer must find it intact.
//
// Caller must hold g.mu.
func (g *Game) finishMoveLifeSequenceLocked(seq moveLifeSequence) error {
	if !seq.hasRoute {
		return nil
	}
	route := seq.route
	if len(route.simultaneousExit) > 0 {
		route.simultaneousExit = append([]Card(nil), route.simultaneousExit...)
	}
	defer g.publishSimultaneousExitLocked(route.simultaneousExit)()
	return g.runRouteTailLocked(&route)
}
