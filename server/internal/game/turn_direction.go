package game

// turn_direction.go — #6223: "Reverse the game's turn order" (Aeon
// Engine).
//
// CR 103.1 has the default turn order begin with the starting player
// and proceed clockwise; CR 101.4 has every "next player in turn order"
// follow that order. Seats are numbered clockwise, so the default walk
// is seat+1. Game.TurnOrderReversed flips it: every reader that walks
// the table in turn order (the turn rotation, priority, the APNAP
// trigger drain, each APNAP fan-out, the next-turn forecast) asks
// SeatInTurnOrderForEffect for the seat k places along instead of
// adding k itself. A second reversal restores the default order.
//
// It is plain data on Game: the zero value is the default order, and
// the snapshot carries it only while set (turnOrderReversed,
// omitempty), so every older snapshot restores unchanged.

// turnOrderStepLocked is +1 for the default (clockwise) order and -1
// once the order has been reversed an odd number of times. Caller must
// hold g.mu.
func (g *Game) turnOrderStepLocked() int {
	if g.TurnOrderReversed {
		return -1
	}
	return 1
}

// SeatInTurnOrderForEffect is the seat index k places after `from` in
// the game's current turn order (k may be 0, negative, or larger than
// the table). It does not skip departed seats: callers walk k = 1, 2,
// ... and pass the ones they do not want, as they always have. -1 when
// the table has no seats. Caller must hold g.mu (read or write).
func (g *Game) SeatInTurnOrderForEffect(from, k int) int {
	n := len(g.Seats)
	if n == 0 {
		return -1
	}
	i := (from + k*g.turnOrderStepLocked()) % n
	if i < 0 {
		i += n
	}
	return i
}

// TurnOrderReversedForEffect reports whether the game's turn order is
// currently the reverse of the default. Caller must hold g.mu.
func (g *Game) TurnOrderReversedForEffect() bool { return g.TurnOrderReversed }

// ReverseTurnOrderForEffect is "reverse the game's turn order". The
// current turn is not changed; the next turn, the next priority pass
// and every APNAP walk from now on go the other way round the table.
// Normal rotation after a run of extra turns resumes from the seat
// whose normal turn they followed (Turn.OrderSeat), in the new
// direction.
//
// Caller must hold g.mu in write mode (the effect_api.go convention).
func (g *Game) ReverseTurnOrderForEffect() {
	g.TurnOrderReversed = !g.TurnOrderReversed
}
