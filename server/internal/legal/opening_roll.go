package legal

// opening_roll.go is ADR 0121 §4: the moves of the opening roll. While
// the roll is open a seat is offered, and only offered:
//
//   - roll_opening, when it is in the current round and has not rolled
//     in it, and nobody has won yet;
//   - when it is the winner, one choose_starting_player per seat still
//     in the game, its own included (CR 103.1: the winner chooses who
//     takes the first turn);
//   - nothing otherwise.
//
// The roll and the chooser's own seat are AlwaysLegal; another seat is
// not. A die is the seat's own: nothing another seat does takes it
// away short of ending the game, and a concession that leaves this seat
// the round's only contender makes it the chooser without making its
// die illegal (RollOpening still rolls it). The chooser can always name
// itself, but another seat can concede between the enumeration and the
// dispatch, and naming it is then refused (ErrPlayerEliminated).
//
// host_roll_remaining is never offered: a bot is never the host
// (ws/host.go). Neither is a table roll (ADR 0121 §5): it changes
// nothing in the game, and a random-tier bot would roll every window.

type chooseStartingPlayerParams struct {
	Seat int `json:"seat"`
}

// openingRollMoves enumerates the seat's opening-roll moves. Caller has
// checked g.OpeningRoll is non-nil and holds the read lock.
func (e *enumerator) openingRollMoves() {
	or := e.g.OpeningRoll
	if or.Chooser < 0 {
		if e.g.OwesOpeningDieLocked(e.p.Seat) {
			e.add(Move{
				Type:        TypeRollOpening,
				Player:      e.seat,
				Kind:        KindOpeningRoll,
				Label:       "Roll a d20 for the first turn",
				AlwaysLegal: true,
			})
		}
		return
	}
	if or.Chooser != e.p.Seat {
		return
	}
	for _, s := range e.g.Seats {
		if s == nil || s.Eliminated {
			continue
		}
		label := s.Name + " goes first"
		if s.Seat == e.p.Seat {
			label = "I go first"
		}
		e.add(Move{
			Type:        TypeChooseStartingPlayer,
			Player:      e.seat,
			Kind:        KindOpeningRoll,
			Label:       label,
			Params:      mustJSON(chooseStartingPlayerParams{Seat: s.Seat}),
			AlwaysLegal: s.Seat == e.p.Seat,
		})
	}
}
