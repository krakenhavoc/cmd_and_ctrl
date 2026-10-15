package game

import "github.com/google/uuid"

// move_extra_turn.go — an extra turn after a replaced move (ADR 0188
// Decision 2, #3560).
//
// Ugin's Nexus: "If Ugin's Nexus would be put into a graveyard from the
// battlefield, instead exile it and take an extra turn after this one."
// The modified event is the exile and then the extra turn, in printed
// order. Taking an extra turn emits EventExtraTurnAdded, and a Replace
// callback must not emit an event (a window is open and an event can
// itself be replaced), so the replacement only DECLARES the turn on the
// move event, exactly as ChangeLifeAfterReplacedMoveForEffect declares a
// life change (ADR 0178). The move's finisher lands the card, then
// queues the turn (runMoveTailLocked), so the log reads exile first and
// extra_turn after it.
//
// Queuing an extra turn never pauses, so unlike the life instructions it
// needs none of moveLifeSequence's continuation machinery: the turns are
// queued as the tail starts and the rest of the tail carries on.

// extraTurnInstruction is one declared extra turn: who takes it and the
// source the log credits. Both are fixed as the replacement applies, so
// a source that leaves afterwards does not change who gets the turn.
type extraTurnInstruction struct {
	Source uuid.UUID
	Player uuid.UUID
}

// TakeExtraTurnAfterReplacedMoveForEffect declares, from a replacement's
// Replace callback, that the modified zone-change event also gives
// player an extra turn after this one once the card has moved ("instead
// exile it and take an extra turn after this one").
//
// The instruction survives a later replacement that redirects or
// cancels the move (CR 614.6: an impossible instruction is ignored, the
// others still happen). Choosing Rest in Peace FIRST is what removes the
// turn: it replaces the move before the Nexus' own replacement gets to
// apply, so the Nexus never declares anything. A move abandoned before
// it happens, because its prompt was taken away, discards the
// instruction (CR 614.7).
//
// Returns ErrInvalidParam for an event that is not a zone change.
//
// Caller must hold g.mu (Replace callbacks always do).
func (g *Game) TakeExtraTurnAfterReplacedMoveForEffect(ev *ReplacementEvent, source, player uuid.UUID) error {
	if ev == nil || ev.Kind != RepEventMove {
		return ErrInvalidParam
	}
	if player == uuid.Nil {
		return nil
	}
	// A fresh backing array: an undo snapshot taken while this event was
	// paused shares the old one (cloneReplacementResume copies too).
	n := len(ev.extraTurnsAfterMove)
	ev.extraTurnsAfterMove = append(ev.extraTurnsAfterMove[:n:n], extraTurnInstruction{Source: source, Player: player})
	return nil
}

// queueExtraTurnsAfterMoveLocked queues each declared extra turn, in
// printed order, and clears the list so a re-entered tail cannot queue
// them twice. A player who has left the game gets none
// (TakeExtraTurnsForEffect refuses them, CR 800.4a).
//
// Caller must hold g.mu.
func (g *Game) queueExtraTurnsAfterMoveLocked(ev *ReplacementEvent) {
	if ev == nil || len(ev.extraTurnsAfterMove) == 0 {
		return
	}
	in := ev.extraTurnsAfterMove
	ev.extraTurnsAfterMove = nil
	for _, t := range in {
		g.TakeExtraTurnsForEffect(t.Player, t.Source, 1)
	}
}
