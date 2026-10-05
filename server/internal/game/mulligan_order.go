package game

import "errors"

// Mulligans in turn order (CR 103.5, issue #2237).
//
// After the starting player is determined, each player in turn order,
// starting with the starting player, decides whether to keep or
// mulligan. Everyone who mulliganed redraws, and the process repeats
// in turn order for those players only. A player who kept is done.
//
// State: Player.MulliganDecided is true for a seat that has answered in
// the current round. The seat that decides next is derived, never
// stored: the first seat from StartingSeat, in seat order, that is not
// eliminated, has not kept and has not answered this round. When no
// such seat is left the round is over; the seats that mulliganed (the
// ones that have not kept) get a fresh round, and when every live seat
// has kept the window closes.
//
// Engine simplification: a mulligan redraws at once instead of after
// the round. The redraw uses the seat's own shuffle stream and nobody
// else's decision reads it, so only the moment the new hand is dealt
// differs from the paper order, not what anyone can know.

// ErrNotYourMulligan is returned by KeepHand and Mulligan when another
// seat is the one deciding (CR 103.5), or this seat has already
// answered in the current round.
var ErrNotYourMulligan = errors.New("game: it is not your turn to decide on your opening hand")

// MulliganDeciderLocked returns the seat that decides next, or -1 when
// no seat has an open decision. Caller must hold g.mu (read is enough); exported for the
// view builder, which already holds it.
func (g *Game) MulliganDeciderLocked() int {
	n := len(g.Seats)
	if !g.MulligansOpen || g.OpeningRoll != nil || n == 0 {
		return -1
	}
	for k := 0; k < n; k++ {
		i := (g.StartingSeat + k) % n
		p := g.Seats[i]
		if p == nil || p.Eliminated || p.HandKept || p.MulliganDecided {
			continue
		}
		return i
	}
	return -1
}

// MulliganDecider is the seat that has to answer keep or mulligan now,
// or -1 when none does (no window, or the opening roll is open). Takes
// the read lock.
func (g *Game) MulliganDecider() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.MulliganDeciderLocked()
}

// settleMulliganLocked advances the mulligan window after any change
// to it: a keep, a mulligan, or a departure. It starts the next round
// for the seats that mulliganed, and closes the window when every live
// seat has kept. Safe to call at any time; it does nothing outside the
// window. Caller must hold g.mu.
func (g *Game) settleMulliganLocked() {
	if g.State != StateActive || !g.MulligansOpen || g.OpeningRoll != nil {
		return
	}
	if g.MulliganDeciderLocked() >= 0 {
		return
	}
	allKept := true
	for _, p := range g.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		if !p.HandKept {
			allKept = false
			p.MulliganDecided = false
		}
	}
	if !allKept {
		// Next round: the seats that mulliganed, in turn order again.
		return
	}
	g.MulligansOpen = false
	// First-step entry happens here, not at Start: the cursor has been
	// parked on Untap with NoPriority since Start, waiting for everyone
	// to commit. Run the hook now so the starting seat's auto-untap
	// fires and the cursor advances to Upkeep. (S13.)
	g.runStepEntryHooksLocked()
}
