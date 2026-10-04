package protocol

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// OpeningRollView is GameView.opening_roll (ADR 0121 §3): the open
// opening roll, present only while it is open. Everything in it is
// public — every die is seen by the whole table the moment it lands —
// so FilterViewFor passes it through to every viewer unchanged.
type OpeningRollView struct {
	// Rounds are the rounds so far: round 1 is every seat, each later
	// round the tied leaders of the one before.
	Rounds []OpeningRollRoundView `json:"rounds"`
	// Chooser is the seat that chooses who takes the first turn,
	// present once one leader remains.
	Chooser *int `json:"chooser,omitempty"`
}

// OpeningRollRoundView is one round: the seats that roll in it, in
// seat order, and their dice in the order rolled.
type OpeningRollRoundView struct {
	Seats []int                `json:"seats"`
	Rolls []OpeningRollDieView `json:"rolls"`
}

// OpeningRollDieView is one opening d20. By is present only when
// somebody other than the seat pressed the button: the host's seat, or
// -1 (game.OpeningRollByAdmin) for the server admin.
type OpeningRollDieView struct {
	Seat   int  `json:"seat"`
	Result int  `json:"result"`
	By     *int `json:"by,omitempty"`
}

// viewOfOpeningRoll projects the window, nil when it is closed.
func viewOfOpeningRoll(or *game.OpeningRoll) *OpeningRollView {
	if or == nil {
		return nil
	}
	out := &OpeningRollView{Rounds: make([]OpeningRollRoundView, len(or.Rounds))}
	for i, r := range or.Rounds {
		round := OpeningRollRoundView{
			Seats: append([]int{}, r.Seats...),
			Rolls: make([]OpeningRollDieView, len(r.Rolls)),
		}
		for j, d := range r.Rolls {
			die := OpeningRollDieView{Seat: d.Seat, Result: d.Result}
			if d.By != d.Seat {
				by := d.By
				die.By = &by
			}
			round.Rolls[j] = die
		}
		out.Rounds[i] = round
	}
	if or.Chooser >= 0 {
		chooser := or.Chooser
		out.Chooser = &chooser
	}
	return out
}
