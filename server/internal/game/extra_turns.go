package game

import "github.com/google/uuid"

// extra_turns.go — CR 500.7 extra turns (ADR 0059 Decision 5, #753).
//
// "Take an extra turn after this one" puts a turn directly after the
// current one. Several of them are taken one at a time, and the most
// recently created is taken first (CR 500.7). So the queue is a STACK:
// TakeExtraTurnsForEffect pushes onto the end of Game.ExtraTurns and
// the rotation seam (beginNextTurnLocked, rotation.go) pops from the
// end. A Time Warp cast during the first of Time Stretch's two turns
// is therefore taken before the second, with no rule of its own.
//
// An extra turn is a turn. It gets its own Turn.Seq, the seat's
// TurnsBegun goes up, and every per-turn reset runs as it begins
// (onTurnBeganLocked), so a planeswalker can be activated again, the
// land drop comes back and "until your next turn" ends. Turn.Round
// does not move: "Turn N" on the board is the round (ADR 0059 owner
// decision 1), and Turn.Extra marks the turn instead.
//
// When the queue is empty, normal rotation resumes from Turn.OrderSeat
// — the seat whose NORMAL turn this is, or last was — so an extra turn
// seat 0 gives seat 2 in a four-player game is followed by seat 1's
// normal turn, not seat 3's.
//
// A queued turn of a player who has left the game doesn't begin
// (CR 800.4k). It is dropped as it would begin, and it still counts
// toward that seat's TurnsBegun (CR 800.4m), exactly as a departed
// seat's passed-over normal turn does.

// ExtraTurn is one queued CR 500.7 extra turn.
type ExtraTurn struct {
	// Ref is the turn's identity, minted from Game.NextExtraRef. It
	// is stable for the life of the queue entry and becomes Turn.ExtraRef
	// when the turn begins, so a delayed trigger can be bound to "that
	// turn" (DelayedTrigger.OnExtraTurn: Final Fortune's end-step loss).
	Ref int `json:"ref"`
	// Seat is the seat index that takes the turn.
	Seat int `json:"seat"`
	// Source is the card whose effect created the turn, for the log.
	Source uuid.UUID `json:"source,omitempty"`
}

// TakeExtraTurnsForEffect gives `player` n extra turns directly after
// the current turn (CR 500.7). It returns their refs in the order they
// will be taken — the first one taken first — and emits one
// EventExtraTurnAdded per turn.
//
// Returns nil, and queues nothing, for n < 1, for a player who is not
// seated and for one who has left the game: an effect that tells a
// departed player to take a turn does nothing (CR 800.4a).
//
// Safe to call from inside a resolution: it only queues. The turn
// begins at the rotation seam once this one ends, never here.
//
// Caller must hold g.mu.
func (g *Game) TakeExtraTurnsForEffect(player, source uuid.UUID, n int) []int {
	if n < 1 {
		return nil
	}
	p := g.playerByIDLocked(player)
	if p == nil || p.Eliminated {
		return nil
	}
	refs := make([]int, n)
	for i := range refs {
		g.NextExtraRef++
		refs[i] = g.NextExtraRef
	}
	// The stack is popped from the end, so the turn taken first is
	// pushed LAST: all n of this effect's turns go ahead of every turn
	// queued earlier (CR 500.7, "most recently created first"), and
	// among themselves they come out in ref order.
	//
	// A fresh slice every time, never an in-place append: RestoreFrom
	// hands the live game the backing array an undo snapshot holds,
	// and writing into its spare capacity would rewrite that snapshot
	// (the trap sweepScopedStaticsLocked documents).
	queue := make([]ExtraTurn, len(g.ExtraTurns), len(g.ExtraTurns)+n)
	copy(queue, g.ExtraTurns)
	for i := n - 1; i >= 0; i-- {
		queue = append(queue, ExtraTurn{Ref: refs[i], Seat: p.Seat, Source: source})
	}
	g.ExtraTurns = queue
	for _, ref := range refs {
		g.EmitEvent(Event{
			Kind:   EventExtraTurnAdded,
			Actor:  player,
			Source: source,
			Amount: ref,
		})
	}
	return append([]int(nil), refs...)
}

// ExtraTurnsQueuedForEffect is a copy of the extra-turn queue in the
// order the turns will be taken, next first. Caller must hold g.mu.
func (g *Game) ExtraTurnsQueuedForEffect() []ExtraTurn {
	return g.extraTurnsInOrderLocked()
}

// ExtraTurnsQueued is the locking form of ExtraTurnsQueuedForEffect.
func (g *Game) ExtraTurnsQueued() []ExtraTurn {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.extraTurnsInOrderLocked()
}

// extraTurnsInOrderLocked reverses the stack into take order.
func (g *Game) extraTurnsInOrderLocked() []ExtraTurn {
	if len(g.ExtraTurns) == 0 {
		return nil
	}
	out := make([]ExtraTurn, 0, len(g.ExtraTurns))
	for i := len(g.ExtraTurns) - 1; i >= 0; i-- {
		out = append(out, g.ExtraTurns[i])
	}
	return out
}

// popExtraTurnLocked takes the next queued extra turn of a seat still
// in the game. Entries of seats that have left are dropped on the way
// (CR 800.4k) and counted toward their TurnsBegun (CR 800.4m). ok is
// false when the queue held no playable turn.
//
// Caller must hold g.mu.
func (g *Game) popExtraTurnLocked() (ExtraTurn, bool) {
	for len(g.ExtraTurns) > 0 {
		last := len(g.ExtraTurns) - 1
		et := g.ExtraTurns[last]
		// Re-sliced to a length the caller can only shrink or copy
		// from: TakeExtraTurnsForEffect never appends in place, so the
		// shared backing array is never written through this slice.
		g.ExtraTurns = g.ExtraTurns[:last:last]
		if len(g.ExtraTurns) == 0 {
			g.ExtraTurns = nil
		}
		if et.Seat < 0 || et.Seat >= len(g.Seats) || g.Seats[et.Seat] == nil {
			continue
		}
		if g.Seats[et.Seat].Eliminated {
			g.noteTurnBegunLocked(et.Seat)
			continue
		}
		if g.extraTurnSkippedLocked(et) {
			continue
		}
		return et, true
	}
	return ExtraTurn{}, false
}

// extraTurnSkippedLocked opens the CR 614 window over a queued extra
// turn that is about to begin and reports whether it settled on a skip
// (CR 614.10): "if an opponent would begin an extra turn, that player
// skips that turn instead" (Trouble in Pairs, #2529).
//
// A skipped turn never begins. It is already off the queue, so nothing
// bound to it can fire — sweepUnreachableBoundTriggersLocked drops a
// delayed trigger tied to its Ref as the next turn starts (CR 614.10a:
// "anything scheduled for a skipped turn won't happen", which is why
// Final Fortune's lose-the-game never comes) — and it does not count
// toward the seat's TurnsBegun, the same call ADR 0059 Decision 1 makes
// for every turn skipped by an effect: that count is the turns that
// BEGAN. A seat that has left the game is dropped before this runs
// (CR 800.4k), and the window never sees it.
//
// The window cannot pause (mustSettleNow): the rotation seam is where
// the next turn is chosen and has no resume. See RepEventExtraTurn.
//
// Caller must hold g.mu.
func (g *Game) extraTurnSkippedLocked(et ExtraTurn) bool {
	ev := &ReplacementEvent{
		Kind:          RepEventExtraTurn,
		Actor:         g.Seats[et.Seat].ID,
		Source:        et.Source,
		ExtraTurnSeat: et.Seat,
		ExtraTurnRef:  et.Ref,
		mustSettleNow: true,
	}
	out, err := g.applyReplacementsLocked(ev)
	defer g.clearReplacementEventLocked(ev.ID)
	// A cancelled event comes back as (nil, nil) — the signal
	// runStepEntryHooksLocked reads the same way. An error is a broken
	// pipeline (mustSettleNow forecloses a pause), and the turn is then
	// taken: weaker than printed for the skipper, never a turn that
	// silently vanishes.
	if err != nil || (out != nil && !out.Canceled) {
		return false
	}
	g.EmitEvent(Event{
		Kind:   EventExtraTurnSkipped,
		Actor:  ev.Actor,
		Source: et.Source,
		Amount: et.Ref,
	})
	return true
}

// extraTurnPendingLocked reports whether `ref` names the turn now in
// progress or one still queued. A delayed trigger bound to any other
// ref can never fire and is swept (sweepUnreachableBoundTriggersLocked).
func (g *Game) extraTurnPendingLocked(ref int) bool {
	if g.Turn.ExtraRef == ref {
		return true
	}
	for _, et := range g.ExtraTurns {
		if et.Ref == ref {
			return true
		}
	}
	return false
}

func cloneExtraTurns(in []ExtraTurn) []ExtraTurn {
	if len(in) == 0 {
		return nil
	}
	return append([]ExtraTurn(nil), in...)
}
