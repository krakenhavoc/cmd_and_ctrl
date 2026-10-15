package game

// next_turn_seat.go — ADR 0188 Decision 9, #3566: who takes the next turn.
//
// The table, the bots and the model prompt all need to know whose turn
// comes after this one. Before this, each guessed: the heuristic's
// nextTurnSeat rotated from ActiveSeat, which is wrong during an extra
// turn (seat 2's extra turn after seat 0's names seat 3, not seat 1),
// and none of them could see a skip charge or a departed seat. The
// engine already decides it at the rotation seam (beginNextTurnLocked);
// this is that decision run as a read-only simulation, so the two cannot
// drift for the cases the records can express.
//
// What the simulation does not see: a replacement skip (Trouble in
// Pairs, Ugin's Nexus), which is decided as the turn would begin, and a
// turn-long control's extra turn (Emrakul), which is queued as the
// current turn ends. Both are rarer than the cases it does cover, and
// each makes the answer wrong by at most the one turn.

// NextTurnSeatForEffect is the seat index whose turn begins after the
// current one, as far as the records can say: the extra-turn stack from
// the top, then normal rotation from Turn.OrderSeat, passing a seat that
// has left (CR 800.4k) and spending each seat's skip charges (CR 614.10)
// in a copy, exactly as beginNextTurnLocked does. ok is false when the
// game is not in play, no seat is left to take a turn, or the cursor
// names no seat.
//
// It writes nothing. Caller must hold g.mu (read or write).
func (g *Game) NextTurnSeatForEffect() (seat int, ok bool) {
	n := len(g.Seats)
	if n == 0 || g.State != StateActive {
		return -1, false
	}
	owed := make(map[int]int)
	total := 0
	for i := range g.ScopedEffects {
		for _, m := range g.ScopedEffects[i].Mods {
			if m.Kind != ModSkipTurns || m.Amount < 1 {
				continue
			}
			for s, p := range g.Seats {
				if p != nil && p.ID == m.Player {
					owed[s] += m.Amount
					total += m.Amount
					break
				}
			}
		}
	}
	// takes reports whether the seat would begin a turn now. A departed
	// seat is passed; a seat that owes a skip spends one charge and is
	// passed.
	takes := func(s int) bool {
		if s < 0 || s >= n || g.Seats[s] == nil || g.Seats[s].Eliminated {
			return false
		}
		if owed[s] > 0 {
			owed[s]--
			return false
		}
		return true
	}
	for _, et := range g.extraTurnsInOrderLocked() {
		if takes(et.Seat) {
			return et.Seat, true
		}
	}
	from := g.Turn.ActiveSeat
	if g.Turn.Extra {
		// Normal rotation resumes from the seat whose normal turn the
		// extra turns followed (beginNextTurnLocked).
		from = g.Turn.OrderSeat
	}
	if from < 0 || from >= n {
		return -1, false
	}
	// Every step takes the turn, passes a departed seat or spends a
	// charge, so n + total + 1 steps always reach a verdict.
	for i := 0; i < n+total+1; i++ {
		from = (from + 1) % n
		if takes(from) {
			return from, true
		}
	}
	return -1, false
}

// NextTurnSeat is the locking form of NextTurnSeatForEffect.
func (g *Game) NextTurnSeat() (int, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.NextTurnSeatForEffect()
}
